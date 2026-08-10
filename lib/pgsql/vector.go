package pgsql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/db/types"
	"github.com/getevo/evo/v2/lib/log"
	"gorm.io/gorm"
)

// vectorExtension is the statement that makes the vector type available. It is
// emitted whenever a vector column is created, never on its own, so that a
// steady state migration stays empty.
//
// It needs the privilege to create an extension. Where that is not granted,
// run it once by hand: IF NOT EXISTS then makes this a no-op.
const vectorExtension = `CREATE EXTENSION IF NOT EXISTS vector;`

// isVectorType reports whether a column type is a pgvector vector, with or
// without a declared width.
func isVectorType(columnType string) bool {
	var t = strings.ToLower(strings.TrimSpace(columnType))
	return t == "vector" || strings.HasPrefix(t, "vector(")
}

// vectorDimensions returns the declared width of a vector type, or zero when
// the type carries none or is not a vector at all.
func vectorDimensions(columnType string) int {
	if !isVectorType(columnType) {
		return 0
	}
	var t = strings.ToLower(strings.TrimSpace(columnType))
	open := strings.Index(t, "(")
	if open < 0 || !strings.HasSuffix(t, ")") {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(t[open+1 : len(t)-1]))
	if err != nil {
		return 0
	}
	return n
}

// tableHasVector reports whether any column of a table is a vector.
func tableHasVector(t pgDdlTable) bool {
	for _, col := range t.Columns {
		if isVectorType(col.Type) {
			return true
		}
	}
	return false
}

// vectorIndex builds the index described by the vector_index tag of a field.
//
//	`gorm:"size:1536;vector_index:hnsw"`          -- HNSW, cosine
//	`gorm:"size:1536;vector_index:hnsw,l2"`       -- HNSW, straight line distance
//	`gorm:"size:1536;vector_index:ivfflat,ip"`    -- IVFFlat, inner product
//	`gorm:"size:1536;vector_index"`               -- HNSW, cosine
//
// The metric decides the operator class, and a query only uses the index when
// it measures distance the same way. Returns nil when the field asks for no
// index, or when it asks for one it cannot have.
func vectorIndex(tableName string, column schema.Column, tag string, hasTag bool) *schema.Index {
	if !hasTag {
		return nil
	}

	var method = "hnsw"
	var metric = types.VectorCosine

	var parts = strings.Split(strings.ToLower(strings.TrimSpace(tag)), ",")
	if len(parts) > 0 && parts[0] != "" {
		method = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
		metric = types.VectorMetric(strings.TrimSpace(parts[1]))
	}

	if method != "hnsw" && method != "ivfflat" {
		log.Errorf("pgsql: unknown vector index method %q on %s.%s, expected hnsw or ivfflat",
			method, tableName, column.Name)
		return nil
	}
	if !metric.Valid() {
		log.Errorf("pgsql: unknown vector index metric %q on %s.%s, expected cosine, l2, ip or l1",
			metric, tableName, column.Name)
		return nil
	}
	// pgvector refuses to index a column that could hold vectors of any width.
	if vectorDimensions(column.Type) == 0 {
		log.Errorf("pgsql: cannot index %s.%s: declare the number of dimensions with the size tag",
			tableName, column.Name)
		return nil
	}

	return &schema.Index{
		Name:    schema.SafeIndexName("vx_"+tableName+"_"+column.Name, 63),
		Columns: schema.Columns{column},
		Method:  method,
		Ops:     metric.OpsClass(),
	}
}

// vectorIndexSQL renders the CREATE INDEX statement of a vector index.
func (p *PGDialect) vectorIndexSQL(index schema.Index, tableName string) string {
	var col = index.Columns[0]
	return fmt.Sprintf(`CREATE INDEX %s ON %s USING %s (%s %s);`,
		p.Quote(index.Name), p.Quote(tableName), index.Method, p.Quote(col.Name), index.Ops)
}

// EnsureVectorExtension creates the vector extension if it is missing. The
// migrator does this on its own when it creates a vector column; this is for
// code that builds its schema some other way, such as a test or a script.
func EnsureVectorExtension(db *gorm.DB) error {
	return db.Exec(vectorExtension).Error
}
