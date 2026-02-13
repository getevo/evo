package pgsql

import (
	"strings"

	"github.com/getevo/evo/v2/lib/db/schema"
)

// zeroDefault returns the appropriate zero-value default for a PG column type.
// Used when adding a NOT NULL column to an existing table.
func (p *PGDialect) zeroDefault(colType string) string {
	t := strings.ToLower(strings.TrimSpace(colType))
	switch {
	case strings.HasPrefix(t, "varchar"), strings.HasPrefix(t, "character varying"),
		t == "text", strings.HasPrefix(t, "char"):
		return "''"
	case t == "boolean", t == "bool":
		return "false"
	case t == "bigint", t == "int8", t == "integer", t == "int4", t == "int",
		t == "smallint", t == "int2":
		return "0"
	case strings.HasPrefix(t, "numeric"), strings.HasPrefix(t, "decimal"),
		t == "real", t == "float4", t == "double precision", t == "float8":
		return "0"
	case t == "timestamp", t == "timestamptz",
		strings.HasPrefix(t, "timestamp"):
		return "CURRENT_TIMESTAMP"
	case t == "date":
		return "CURRENT_DATE"
	case t == "jsonb", t == "json":
		return "'{}'"
	default:
		// For enum types and unknown types, skip zero default
		return ""
	}
}

// normalizeType normalizes PostgreSQL type names for comparison.
func (p *PGDialect) normalizeType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	// Normalize PG aliases
	switch t {
	case "int4", "integer":
		return "int"
	case "int8":
		return "bigint"
	case "int2":
		return "smallint"
	case "float4":
		return "real"
	case "float8":
		return "double precision"
	case "bool":
		return "boolean"
	case "timestamptz":
		return "timestamptz"
	case "tinyint(1)":
		return "boolean"
	case "bigint(20)":
		return "bigint"
	case "decimal":
		return "numeric"
	}
	// Normalize decimal(p,s) -> numeric(p,s)
	if strings.HasPrefix(t, "decimal(") {
		return "numeric" + t[7:]
	}
	return t
}

// normalizeDefault normalizes default values for comparison.
func (p *PGDialect) normalizeDefault(d string) string {
	d = strings.TrimSpace(d)
	// Strip PG type casts like 'value'::character varying, 'a'::enum_type, etc.
	if idx := strings.Index(d, "::"); idx > 0 {
		d = d[:idx]
	}
	d = schema.TrimQuotes(d)
	d = strings.ToLower(d)
	// Normalize PG-style nextval defaults
	if strings.HasPrefix(d, "nextval(") {
		return ""
	}
	// Normalize boolean defaults
	switch d {
	case "1", "true", "t", "'t'":
		return "true"
	case "0", "false", "f", "'f'":
		return "false"
	}
	// Normalize timestamp variants
	switch d {
	case "current_timestamp", "current_timestamp()", "now()":
		return "current_timestamp"
	case "null", "":
		return ""
	}
	return d
}

// needsUsing determines if a type conversion requires a USING clause.
func needsUsing(fromType, toType string) bool {
	from := strings.ToLower(fromType)
	to := strings.ToLower(toType)
	if from == to {
		return false
	}
	// Common conversions that need USING
	if (from == "text" || strings.HasPrefix(from, "varchar")) && (to == "int" || to == "bigint" || to == "boolean") {
		return true
	}
	if (from == "int" || from == "bigint") && to == "boolean" {
		return true
	}
	if from == "boolean" && (to == "int" || to == "bigint") {
		return true
	}
	return false
}

// parseEnumValues returns the individual unquoted values from an enum definition.
// Input: "enum('a','b','c')" -> ["a", "b", "c"]
func parseEnumValues(enumType string) []string {
	s := enumType
	idx := strings.Index(strings.ToLower(s), "enum(")
	if idx >= 0 {
		s = s[idx+5:]
	}
	s = strings.TrimSuffix(s, ")")
	s = strings.TrimSpace(s)
	var vals []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "'\"")
		if part != "" {
			vals = append(vals, part)
		}
	}
	return vals
}

// extractEnumValues extracts and normalizes enum values for PostgreSQL.
// Input: enum('a','b','c') or enum("a","b")
// Output: 'a','b','c' (with single quotes for PG)
func extractEnumValues(enumType string) string {
	s := enumType
	// Remove "enum(" prefix and ")" suffix
	idx := strings.Index(strings.ToLower(s), "enum(")
	if idx >= 0 {
		s = s[idx+5:]
	}
	s = strings.TrimSuffix(s, ")")
	s = strings.TrimSpace(s)
	// Normalize quotes to single quotes
	s = strings.ReplaceAll(s, `"`, `'`)
	return s
}

// getStringPtr safely dereferences a string pointer and trims quotes.
func getStringPtr(v *string) string {
	if v == nil {
		return ""
	}
	return schema.TrimQuotes(*v)
}
