package schema

import (
	"fmt"
	"gorm.io/gorm"

	"github.com/getevo/evo/v2/lib/db/schema/ddl"
	"github.com/getevo/evo/v2/lib/db/schema/table"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/version"
	"reflect"
	"strings"
)

var migrations []interface{}

const null = "NULL"

func GetMigrationScript(db *gorm.DB) []string {
	var queries []string

	var database = ""
	db.Raw("SELECT DATABASE();").Scan(&database)
	var is table.Tables
	db.Where(table.Table{Database: database}).Find(&is)

	var columns table.Columns
	db.Where(table.Table{Database: database}).Order("TABLE_NAME ASC,ORDINAL_POSITION ASC").Find(&columns)

	var tb *table.Table
	for idx, _ := range columns {
		if tb == nil || columns[idx].Table != tb.Table {
			tb = is.GetTable(columns[idx].Table)
			if tb == nil {
				continue
			}
		}
		tb.Columns = append(tb.Columns, columns[idx])
	}

	var istats []table.IndexStat
	db.Where(table.IndexStat{Database: database}).Order("TABLE_NAME ASC, SEQ_IN_INDEX ASC").Find(&istats)

	var indexMap = map[string]table.Index{}
	for _, item := range istats {
		if item.Name == "PRIMARY" {
			continue
		}
		if _, ok := indexMap[item.Table+item.Name]; !ok {
			indexMap[item.Table+item.Name] = table.Index{
				Name:   item.Name,
				Table:  item.Table,
				Unique: !item.NonUnique,
			}
		}
		var m = indexMap[item.Table+item.Name]
		var c = is.GetTable(item.Table).Columns.GetColumn(item.ColumnName)
		m.Columns = append(m.Columns, *c)
		indexMap[item.Table+item.Name] = m
	}
	for key, item := range indexMap {
		is.GetTable(item.Table).Indexes = append(is.GetTable(item.Table).Indexes, indexMap[key])
	}

	for idx, el := range migrations {
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
			log.Fatal(err)
		}

		if obj, ok := stmt.Model.(interface{ TableName() string }); ok {
			if strings.HasPrefix(obj.TableName(), "information_schema.") {
				continue
			}
		}

		if stmt.Schema == nil {
			log.Fatal(fmt.Errorf("invalid schema for %s", reflect.TypeOf(el)))
		}

		//check if table exists
		var table = is.GetTable(stmt.Schema.Table)

		var q []string
		if table != nil {
			table.Model = migrations[idx]
			table.Reflect = reflect.ValueOf(table.Model)
			q = ddl.FromStatement(stmt).GetDiff(*table)

		} else {
			q = ddl.FromStatement(stmt).GetCreateQuery()
		}
		if len(q) > 0 {
			queries = append(queries, "\r\n\r\n-- Migrate Table:"+stmt.Schema.Table)
			queries = append(queries, q...)
		}
		if caller, ok := el.(interface{ Migration() []Migration }); ok {
			var currentVersion = "0.0.0"
			if table != nil {
				db.Raw("SELECT table_comment FROM INFORMATION_SCHEMA.TABLES  WHERE table_schema=?  AND table_name=?", database, table.Table).Scan(&currentVersion)
			}
			var buff []string
			var ptr = "0.0.0"
			for _, item := range caller.Migration() {
				if version.Compare(currentVersion, item.Version, "<") {
					if version.Compare(ptr, item.Version, "<=") {
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
				queries = append(queries, "\r\n\r\n-- Migrate "+stmt.Schema.Table+".Migrate:")
				queries = append(queries, buff...)
				queries = append(queries, "ALTER TABLE `"+stmt.Schema.Table+"`  COMMENT '"+ptr+"';")
			}
		}

	}

	return queries
}

func DoMigration(db *gorm.DB) error {
	//check if tidb
	var tidbMultiStatementMode = ""
	db.Raw("SELECT @@GLOBAL.tidb_multi_statement_mode").Scan(&tidbMultiStatementMode)
	if tidbMultiStatementMode != "" {
		// enable possibility to run multiple queries at once. BEWARE: DONT LEAVE IT ON FOR SECURITY MEASUREMENTS
		db.Exec("SET tidb_multi_statement_mode='ON';")
		defer func() {
			// back to original value
			db.Exec("SET tidb_multi_statement_mode='" + tidbMultiStatementMode + "';")
		}()
	}
	var err error
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, query := range GetMigrationScript(db) {
			if !strings.HasPrefix(query, "--") {
				err = tx.Debug().Exec(query).Error
				if err != nil {
					return err
				}
			}
		}

		return nil

	})
	return err
}

type Migration struct {
	Version string
	Query   string
}
