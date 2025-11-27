package schema

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/version"
	"gorm.io/gorm"

	"github.com/getevo/evo/v2/lib/db/schema/ddl"
	"github.com/getevo/evo/v2/lib/db/schema/table"
	"github.com/getevo/evo/v2/lib/log"
	"reflect"
	"strings"
)

var migrations []any

const null = "NULL"

func GetMigrationScript(db *gorm.DB) []string {
	var queries []string

	var database = ""
	var engine string
	db.Raw("SELECT DATABASE();").Scan(&database)
	db.Raw("SELECT VERSION();").Scan(&engine)
	engine = strings.ToLower(engine)
	switch {
	case strings.Contains(engine, "mariadb"):
		ddl.Engine = "mariadb"
	case strings.Contains(engine, "mysql"):
		ddl.Engine = "mysql"
	}

	var is table.Tables
	db.Raw(`SELECT CCSA.character_set_name  AS 'TABLE_CHARSET',T.* FROM information_schema.TABLES T, information_schema.COLLATION_CHARACTER_SET_APPLICABILITY CCSA WHERE CCSA.collation_name = T.table_collation AND T.table_schema = ?`, database).Scan(&is)

	var columns table.Columns
	db.Where(table.Table{Database: database}).Order("TABLE_NAME ASC,ORDINAL_POSITION ASC").Find(&columns)

	var constraints []table.Constraint
	db.Where(table.Constraint{Database: database}).Find(&constraints)

	var tb *table.Table
	for idx, _ := range columns {
		if tb == nil || columns[idx].Table != tb.Table {
			tb = is.GetTable(columns[idx].Table)
			if tb == nil {
				continue
			}
		}
		if columns[idx].ColumnKey == "PRI" {
			tb.PrimaryKey = append(tb.Columns, columns[idx])
		}
		tb.Columns = append(tb.Columns, columns[idx])
	}

	var istats []table.IndexStat
	db.Where(table.IndexStat{Database: database}).Order("TABLE_NAME ASC, SEQ_IN_INDEX ASC").Find(&istats)
	var tail []string
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

		//q = append(q, ddl.FromStatement(stmt).Constrains(constraints, is)...)
		tail = append(tail, ddl.FromStatement(stmt).Constrains(constraints, is)...)
		if len(q) > 0 {
			queries = append(queries, "\r\n\r\n-- Migrate Model: "+stmt.Schema.ModelType.PkgPath()+"."+stmt.Schema.ModelType.Name()+"("+stmt.Schema.Table+")")
			queries = append(queries, q...)
		}
		if caller, ok := el.(interface {
			Migration(version string) []Migration
		}); ok {
			var currentVersion = "0.0.0"
			if table != nil {
				db.Raw("SELECT table_comment FROM INFORMATION_SCHEMA.TABLES  WHERE table_schema=?  AND table_name=?", database, table.Table).Scan(&currentVersion)
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
				queries = append(queries, "\r\n\r\n-- Migrate "+stmt.Schema.Table+".Migrate:")
				queries = append(queries, buff...)
				queries = append(queries, "ALTER TABLE `"+stmt.Schema.Table+"`  COMMENT '"+ptr+"';")
			}
		}

	}
	queries = append(queries, tail...)
	return queries
}

func DoMigration(db *gorm.DB) error {
	var err error
	//err = db.Transaction(func(tx *gorm.DB) error {
	for _, query := range GetMigrationScript(db) {
		query = strings.TrimSpace(query)
		if !strings.HasPrefix(strings.TrimSpace(query), "--") && strings.TrimSpace(query) != "" {
			err = db.Debug().Exec(query).Error
			if err != nil {
				log.Error(err)
			}
		} else {
			fmt.Println(query)
		}
	}

	//return nil

	//})
	return err
}

type Migration struct {
	Version string
	Query   string
}
