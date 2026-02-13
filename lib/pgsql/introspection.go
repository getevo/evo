package pgsql

import (
	"gorm.io/gorm"
)

// introspectTables retrieves table metadata from the database.
func (p *PGDialect) introspectTables(db *gorm.DB, database string) pgRemoteTables {
	var is pgRemoteTables
	db.Raw(`
		SELECT table_catalog   AS table_schema,
		       table_name      AS table_name,
		       table_type      AS table_type,
		       ''              AS engine,
		       ''              AS table_collation,
		       ''              AS table_charset
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_catalog = ?
		  AND table_type = 'BASE TABLE'
	`, p.Schema(), database).Scan(&is)
	return is
}

// introspectColumns retrieves column metadata from the database.
func (p *PGDialect) introspectColumns(db *gorm.DB, database string) pgRemoteColumns {
	var columns pgRemoteColumns
	db.Raw(`
		SELECT c.table_catalog                AS table_schema,
		       c.table_name                   AS table_name,
		       c.column_name                  AS column_name,
		       c.ordinal_position             AS ordinal_position,
		       c.column_default               AS column_default,
		       c.is_nullable                  AS is_nullable,
		       c.udt_name                     AS data_type,
		       CASE
		           WHEN c.data_type = 'USER-DEFINED' THEN c.udt_name
		           WHEN c.udt_name = 'varchar' AND c.character_maximum_length IS NOT NULL
		               THEN 'varchar(' || c.character_maximum_length || ')'
		           WHEN c.udt_name = 'bpchar' AND c.character_maximum_length IS NOT NULL
		               THEN 'char(' || c.character_maximum_length || ')'
		           WHEN c.udt_name = 'numeric' AND c.numeric_precision IS NOT NULL
		               THEN 'decimal(' || c.numeric_precision || ',' || COALESCE(c.numeric_scale, 0) || ')'
		           WHEN c.udt_name = 'int4' THEN 'int'
		           WHEN c.udt_name = 'int8' THEN 'bigint'
		           WHEN c.udt_name = 'int2' THEN 'smallint'
		           WHEN c.udt_name = 'float4' THEN 'real'
		           WHEN c.udt_name = 'float8' THEN 'double precision'
		           WHEN c.udt_name = 'bool' THEN 'boolean'
		           WHEN c.udt_name = 'timestamptz' THEN 'timestamptz'
		           WHEN c.udt_name = 'timestamp' THEN 'timestamp'
		           WHEN c.udt_name = 'text' THEN 'text'
		           WHEN c.udt_name = 'jsonb' THEN 'jsonb'
		           WHEN c.udt_name = 'json' THEN 'json'
		           ELSE c.udt_name
		       END                            AS column_type,
		       c.character_maximum_length      AS character_maximum_length,
		       c.numeric_precision             AS numeric_precision,
		       c.numeric_scale                 AS numeric_scale,
		       c.datetime_precision            AS datetime_precision,
		       ''                              AS character_set_name,
		       ''                              AS collation_name,
		       CASE
		           WHEN pk.column_name IS NOT NULL THEN 'PRI'
		           ELSE ''
		       END                            AS column_key,
		       CASE
		           WHEN c.column_default LIKE 'nextval(%%' THEN 'auto_increment'
		           ELSE ''
		       END                            AS extra,
		       COALESCE(col_description(cls.oid, c.ordinal_position), '') AS column_comment
		FROM information_schema.columns c
		LEFT JOIN pg_class cls
		       ON cls.relname = c.table_name AND cls.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = ?)
		LEFT JOIN (
		    SELECT ku.table_name, ku.column_name
		    FROM information_schema.key_column_usage ku
		    JOIN information_schema.table_constraints tc
		        ON ku.constraint_name = tc.constraint_name AND ku.table_schema = tc.table_schema
		    WHERE tc.constraint_type = 'PRIMARY KEY'
		      AND ku.table_schema = ?
		) pk ON pk.table_name = c.table_name AND pk.column_name = c.column_name
		WHERE c.table_schema = ?
		  AND c.table_catalog = ?
		ORDER BY c.table_name, c.ordinal_position
	`, p.Schema(), p.Schema(), p.Schema(), database).Scan(&columns)
	return columns
}

// introspectConstraints retrieves foreign key constraint metadata from the database.
func (p *PGDialect) introspectConstraints(db *gorm.DB) []pgConstraint {
	var constraints []pgConstraint
	db.Raw(`
		SELECT con.conname                              AS constraint_name,
		       cl.relname                               AS table_name,
		       att.attname                              AS column_name,
		       ref_cl.relname                           AS referenced_table_name,
		       ref_att.attname                          AS referenced_column_name
		FROM pg_constraint con
		JOIN pg_class cl ON con.conrelid = cl.oid
		JOIN pg_namespace ns ON cl.relnamespace = ns.oid
		JOIN pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ANY(con.conkey)
		JOIN pg_class ref_cl ON con.confrelid = ref_cl.oid
		JOIN pg_attribute ref_att ON ref_att.attrelid = con.confrelid AND ref_att.attnum = ANY(con.confkey)
		WHERE con.contype = 'f'
		  AND ns.nspname = ?
	`, p.Schema()).Scan(&constraints)
	return constraints
}

// introspectIndexes retrieves index metadata from the database.
func (p *PGDialect) introspectIndexes(db *gorm.DB, database string, is pgRemoteTables) {
	var istats []pgRemoteIndexStat
	db.Raw(`
		SELECT ?                             AS table_schema,
		       t.relname                     AS table_name,
		       NOT ix.indisunique            AS non_unique,
		       i.relname                     AS index_name,
		       a.attname                     AS column_name
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_namespace ns ON t.relnamespace = ns.oid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE ns.nspname = ?
		  AND NOT ix.indisprimary
		ORDER BY t.relname, array_position(ix.indkey, a.attnum)
	`, database, p.Schema()).Scan(&istats)

	var indexMap = map[string]pgRemoteIndex{}
	for _, item := range istats {
		if item.Name == "PRIMARY" {
			continue
		}
		if _, ok := indexMap[item.Table+item.Name]; !ok {
			indexMap[item.Table+item.Name] = pgRemoteIndex{
				Name:   item.Name,
				Table:  item.Table,
				Unique: !item.NonUnique,
			}
		}
		var m = indexMap[item.Table+item.Name]
		tbl := is.GetTable(item.Table)
		if tbl == nil {
			continue
		}
		var c = tbl.Columns.GetColumn(item.ColumnName)
		if c == nil {
			continue
		}
		m.Columns = append(m.Columns, *c)
		indexMap[item.Table+item.Name] = m
	}
	for key, item := range indexMap {
		tbl := is.GetTable(item.Table)
		if tbl != nil {
			tbl.Indexes = append(tbl.Indexes, indexMap[key])
		}
	}
}

// assembleColumns associates columns with their tables and identifies primary keys.
func (p *PGDialect) assembleColumns(columns pgRemoteColumns, is pgRemoteTables) {
	var tb *pgRemoteTable
	for idx := range columns {
		if tb == nil || columns[idx].Table != tb.Table {
			tb = is.GetTable(columns[idx].Table)
			if tb == nil {
				continue
			}
		}
		if columns[idx].ColumnKey == "PRI" {
			tb.PrimaryKey = append(tb.PrimaryKey, columns[idx])
		}
		tb.Columns = append(tb.Columns, columns[idx])
	}
}
