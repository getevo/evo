package schema

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/version"
	"gorm.io/gorm"
)

var OnBeforeMigration []func(db *gorm.DB)
var OnAfterMigration []func(db *gorm.DB)

var migrations []any

// ResetMigrations clears the registered migrations and models (used for testing).
func ResetMigrations() {
	migrations = nil
	Models = nil
	database = ""
}

const null = "NULL"

func GetMigrationScript(db *gorm.DB) []string {
	var queries []string

	// Initialize dialect if needed
	d := GetDialect()
	if d == nil {
		d = InitDialect(db)
	}

	// Get current database
	var database = d.GetCurrentDatabase(db)

	// Parse all registered models into GORM statements
	var stmts []*gorm.Statement
	var modelSlice []any
	for _, el := range migrations {
		var ref = reflect.ValueOf(el)
		for {
			if ref.Kind() == reflect.Ptr {
				ref = ref.Elem()
			} else {
				break
			}
		}
		if ref.Kind() != reflect.Struct {
			continue
		}

		var stmt = db.Model(el).Statement
		var err = stmt.Parse(el)
		if err != nil {
			log.Error("failed to parse model", "model", reflect.TypeOf(el), "error", err)
			continue
		}

		if obj, ok := stmt.Model.(interface{ TableName() string }); ok {
			if strings.HasPrefix(obj.TableName(), "information_schema.") {
				continue
			}
		}

		if stmt.Schema == nil {
			log.Error("invalid schema", "model", reflect.TypeOf(el))
			continue
		}

		stmts = append(stmts, stmt)
		modelSlice = append(modelSlice, el)
	}

	// Delegate all introspection + DDL generation to the dialect
	result := d.GenerateMigration(db, database, stmts, modelSlice)

	queries = append(queries, result.Queries...)

	// Version-based migrations (shared across dialects)
	for idx, el := range modelSlice {
		if caller, ok := el.(interface {
			Migration(version string) []Migration
		}); ok {
			tableName := stmts[idx].Schema.Table
			var currentVersion = "0.0.0"
			if result.TableExists[tableName] {
				currentVersion = d.GetTableVersion(db, database, tableName)
			}
			var buff []string
			var ptr = "0.0.0"
			for _, item := range caller.Migration(currentVersion) {
				if item.Version == "*" || version.Compare(currentVersion, item.Version, "<") {
					if item.Version == "*" {
						ptr = currentVersion
					} else if version.Compare(ptr, item.Version, "<=") {
						ptr = item.Version
					}
					item.Query = strings.TrimSpace(item.Query)
					if !strings.HasSuffix(item.Query, ";") {
						item.Query += ";"
					}
					buff = append(buff, item.Query)
				}
			}

			if len(buff) > 0 {
				queries = append(queries, "\r\n\r\n-- Migrate "+tableName+".Migrate:")
				queries = append(queries, buff...)
				queries = append(queries, d.SetTableVersionSQL(tableName, ptr))
			}
		}
	}

	queries = append(queries, result.Tail...)
	return queries
}

// ComputeSchemaHash builds a deterministic hash from all registered GORM models.
// The hash is dialect-aware (uses Dialector.DataTypeOf) so changes in column
// types, defaults, sizes, etc. produce a different hash.
func ComputeSchemaHash(db *gorm.DB) string {
	// Build a sorted list of table canonical strings
	type tableEntry struct {
		tableName string
		canonical string
	}
	var entries []tableEntry

	for _, el := range migrations {
		ref := reflect.ValueOf(el)
		for ref.Kind() == reflect.Ptr {
			ref = ref.Elem()
		}
		if ref.Kind() != reflect.Struct {
			continue
		}

		stmt := db.Model(el).Statement
		if err := stmt.Parse(el); err != nil {
			continue
		}
		if stmt.Schema == nil {
			continue
		}
		if obj, ok := stmt.Model.(interface{ TableName() string }); ok {
			if strings.HasPrefix(obj.TableName(), "information_schema.") {
				continue
			}
		}

		tableName := stmt.Schema.Table

		// Sort fields by DBName for determinism
		type fieldInfo struct {
			dbName string
			line   string
		}
		var fields []fieldInfo
		for _, f := range stmt.Schema.Fields {
			if f.IgnoreMigration || f.DBName == "" {
				continue
			}
			dataType := stmt.Dialector.DataTypeOf(f)
			notNull := "true"
			if f.FieldType.Kind() == reflect.Ptr {
				notNull = "false"
			}
			if _, ok := f.TagSettings["NULLABLE"]; ok {
				notNull = "false"
			}
			line := fmt.Sprintf("%s|%s|%s|%d|%d|%d|%s|%s",
				tableName, f.DBName, dataType,
				f.Size, f.Precision, f.Scale,
				notNull, f.DefaultValue,
			)
			fields = append(fields, fieldInfo{dbName: f.DBName, line: line})
		}
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].dbName < fields[j].dbName
		})

		var buf strings.Builder
		for _, fi := range fields {
			buf.WriteString(fi.line)
			buf.WriteByte('\n')
		}
		entries = append(entries, tableEntry{tableName: tableName, canonical: buf.String()})
	}

	// Sort by table name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].tableName < entries[j].tableName
	})

	var combined strings.Builder
	for _, e := range entries {
		combined.WriteString(e.canonical)
	}

	return Generate32CharHash(combined.String())
}

// canSkipMigration checks if a successful migration with the same hash already exists.
func canSkipMigration(db *gorm.DB, hash string) bool {
	var id int64
	db.Raw("SELECT id FROM schema_migration WHERE hash = ? AND status = 'success' ORDER BY id DESC LIMIT 1", hash).Scan(&id)
	return id > 0
}

// recordMigration inserts a row into the schema_migration history table.
func recordMigration(db *gorm.DB, hash, status string, executedQueries int, errorMessage string) {
	d := GetDialect()
	now := time.Now().Format("2006-01-02 15:04:05")
	var errMsg *string
	if errorMessage != "" {
		errMsg = &errorMessage
	}
	var err error
	if d != nil && d.Name() == "postgres" {
		err = db.Exec(`INSERT INTO "schema_migration" ("hash","status","executed_queries","error_message","created_at") VALUES (?,?,?,?,?)`,
			hash, status, executedQueries, errMsg, now).Error
	} else {
		err = db.Exec("INSERT INTO `schema_migration` (`hash`,`status`,`executed_queries`,`error_message`,`created_at`) VALUES (?,?,?,?,?)",
			hash, status, executedQueries, errMsg, now).Error
	}
	if err != nil {
		log.Error("failed to record migration history", "error", err)
	}
}

func DoMigration(db *gorm.DB) error {
	d := GetDialect()
	if d == nil {
		d = InitDialect(db)
	}

	// Bootstrap the history table
	if err := d.BootstrapHistoryTable(db); err != nil {
		log.Error("failed to bootstrap schema_migration table", "error", err)
		return err
	}

	// Acquire advisory lock
	if err := d.AcquireMigrationLock(db); err != nil {
		log.Error("failed to acquire migration lock", "error", err)
		return err
	}
	defer d.ReleaseMigrationLock(db)

	// Compute schema hash and check if we can skip
	hash := ComputeSchemaHash(db)
	if !args.Exists("--migration-force") && canSkipMigration(db, hash) {
		log.Info("schema unchanged (hash: " + hash + "), skipping migration")
		return nil
	}

	// Get migration queries
	migrationQueries := GetMigrationScript(db)
	if len(migrationQueries) == 0 {
		recordMigration(db, hash, "success", 0, "")
		log.Info("no migration queries needed, recorded hash: " + hash)
		return nil
	}

	// Execute
	for _, fn := range OnBeforeMigration {
		fn(db)
	}

	var errs []error
	var executed int
	for _, query := range migrationQueries {
		query = strings.TrimSpace(query)
		if !strings.HasPrefix(query, "--") && query != "" {
			if err := db.Debug().Exec(query).Error; err != nil {
				log.Error(err)
				errs = append(errs, err)
			}
			executed++
		} else {
			fmt.Println(query)
		}
	}

	for _, fn := range OnAfterMigration {
		fn(db)
	}

	joinedErr := errors.Join(errs...)
	if joinedErr != nil {
		recordMigration(db, hash, "failed", executed, joinedErr.Error())
	} else {
		recordMigration(db, hash, "success", executed, "")
	}

	return joinedErr
}

// DryRunMigration prints the DDL that would be executed without actually running it.
func DryRunMigration(db *gorm.DB) []string {
	queries := GetMigrationScript(db)
	if len(queries) == 0 {
		fmt.Println("-- No migration queries to execute.")
		return nil
	}
	fmt.Println("-- Migration dry-run: the following queries would be executed:")
	fmt.Println()
	for _, q := range queries {
		fmt.Println(q)
	}
	return queries
}

// DumpSchema prints the full CREATE TABLE DDL for all registered models
// by passing an empty database name so the dialect sees no existing tables.
func DumpSchema(db *gorm.DB) []string {
	d := GetDialect()
	if d == nil {
		d = InitDialect(db)
	}

	// Parse all registered models
	var stmts []*gorm.Statement
	var modelSlice []any
	for _, el := range migrations {
		ref := reflect.ValueOf(el)
		for ref.Kind() == reflect.Ptr {
			ref = ref.Elem()
		}
		if ref.Kind() != reflect.Struct {
			continue
		}

		stmt := db.Model(el).Statement
		if err := stmt.Parse(el); err != nil {
			log.Error("failed to parse model", "error", err)
			continue
		}
		if obj, ok := stmt.Model.(interface{ TableName() string }); ok {
			if strings.HasPrefix(obj.TableName(), "information_schema.") {
				continue
			}
		}
		if stmt.Schema == nil {
			continue
		}
		stmts = append(stmts, stmt)
		modelSlice = append(modelSlice, el)
	}

	// Pass empty database name so the dialect introspects nothing (no existing tables)
	result := d.GenerateMigration(db, "", stmts, modelSlice)

	var all []string
	all = append(all, result.Queries...)
	all = append(all, result.Tail...)

	if len(all) == 0 {
		fmt.Println("-- No models registered.")
		return nil
	}

	fmt.Println("-- Schema dump: full CREATE TABLE DDL for all registered models")
	fmt.Println()
	for _, q := range all {
		fmt.Println(q)
	}
	return all
}

type Migration struct {
	Version string
	Query   string
}
