package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Vector is an embedding: a fixed length list of 32 bit floats, stored in a
// PostgreSQL column of the pgvector `vector` type.
//
// Declare the number of dimensions with the size tag. It is part of the column
// type, and an index cannot be built without it:
//
//	type Document struct {
//	    ID        uint         `gorm:"primaryKey;autoIncrement"`
//	    Body      string       `gorm:"type:text"`
//	    Embedding types.Vector `gorm:"size:1536;vector_index:hnsw"`
//	}
//
// A vector column is nullable unless the model says otherwise, because the
// embedding of a row is usually computed after the row exists.
//
// On MySQL and SQLite the same field is stored as JSON and text respectively.
// The wire format is identical — "[1,2,3]" is both valid pgvector input and
// valid JSON — so a model is portable, but only PostgreSQL can index it or
// measure distance in the database.
type Vector []float32

// Scan reads a vector from the database. pgvector renders it as "[1,2,3]",
// which is also what the JSON and text fallbacks store.
func (v *Vector) Scan(src any) error {
	switch value := src.(type) {
	case nil:
		*v = nil
		return nil
	case []byte:
		return v.parse(string(value))
	case string:
		return v.parse(value)
	default:
		return fmt.Errorf("cannot scan %T into types.Vector", src)
	}
}

// Value writes the vector in the format every supported database accepts. An
// empty vector is stored as NULL: a dimensioned pgvector column rejects a
// zero length value, so "no embedding yet" has to be the null, not "[]".
func (v Vector) Value() (driver.Value, error) {
	if len(v) == 0 {
		return nil, nil
	}
	return v.String(), nil
}

// String renders the vector the way pgvector reads and writes it.
func (v Vector) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(f), 'g', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

// Dimensions reports the length of the vector.
func (v Vector) Dimensions() int { return len(v) }

func (v *Vector) parse(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		*v = nil
		return nil
	}
	if !strings.HasPrefix(raw, "[") || !strings.HasSuffix(raw, "]") {
		return fmt.Errorf("malformed vector %q: expected [x,y,z]", raw)
	}
	raw = strings.TrimSpace(raw[1 : len(raw)-1])
	if raw == "" {
		*v = Vector{}
		return nil
	}

	var parts = strings.Split(raw, ",")
	var out = make(Vector, 0, len(parts))
	for _, part := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return fmt.Errorf("malformed vector element %q: %w", part, err)
		}
		out = append(out, float32(f))
	}
	*v = out
	return nil
}

// GormDataType is the logical type, which the migrator matches on.
func (Vector) GormDataType() string { return "vector" }

// GormDBDataType is the column type. The dimensions come from the size tag;
// without one the column accepts any width, which pgvector allows but cannot
// index.
func (Vector) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		if field != nil && field.Size > 0 {
			return "vector(" + strconv.Itoa(field.Size) + ")"
		}
		return "vector"
	case "mysql":
		return "JSON"
	case "sqlite":
		return "TEXT"
	}
	return "TEXT"
}

// VectorMetric selects how the distance between two vectors is measured. The
// choice decides both the operator used in a query and the operator class an
// index has to be built with, so a query and its index have to agree.
type VectorMetric string

const (
	// VectorCosine measures the angle between two vectors and ignores their
	// magnitude. It is the usual choice for text embeddings.
	VectorCosine VectorMetric = "cosine"
	// VectorL2 measures straight line distance.
	VectorL2 VectorMetric = "l2"
	// VectorInnerProduct measures the negative inner product, which ranks
	// magnitude as well as direction.
	VectorInnerProduct VectorMetric = "ip"
	// VectorL1 measures taxicab distance. Requires pgvector 0.7 or newer.
	VectorL1 VectorMetric = "l1"
)

// operator returns the pgvector distance operator of the metric.
func (m VectorMetric) operator() string {
	switch m {
	case VectorL2:
		return "<->"
	case VectorInnerProduct:
		return "<#>"
	case VectorL1:
		return "<+>"
	default:
		return "<=>"
	}
}

// OpsClass returns the pgvector index operator class of the metric, which is
// what an HNSW or IVFFlat index has to be built with for a query using this
// metric to be able to use it.
func (m VectorMetric) OpsClass() string {
	switch m {
	case VectorL2:
		return "vector_l2_ops"
	case VectorInnerProduct:
		return "vector_ip_ops"
	case VectorL1:
		return "vector_l1_ops"
	default:
		return "vector_cosine_ops"
	}
}

// Valid reports whether the metric is one this package knows.
func (m VectorMetric) Valid() bool {
	switch m {
	case VectorCosine, VectorL2, VectorInnerProduct, VectorL1:
		return true
	}
	return false
}

// Distance builds the distance between a vector column and a value, as an
// expression usable anywhere GORM takes one — a select list, an order, a
// having clause:
//
//	db.Order(clause.OrderBy{Expression: types.Distance("embedding", query, types.VectorCosine)})
//
// The parameter is cast explicitly because PostgreSQL cannot infer the type of
// a bare placeholder on the right of a distance operator.
func Distance(column string, v Vector, metric VectorMetric) clause.Expr {
	return clause.Expr{
		SQL:  fmt.Sprintf("%s %s ?::vector", quoteVectorColumn(column), metric.operator()),
		Vars: []any{v.String()},
	}
}

// Nearest is a scope that orders rows by their distance to a vector, closest
// first:
//
//	db.Model(&Document{}).
//	    Scopes(types.Nearest("embedding", query, types.VectorCosine)).
//	    Limit(5).Find(&docs)
//
// An index built with the matching operator class is used when there is one;
// without one the answer is the same but the whole table is read.
func Nearest(column string, v Vector, metric VectorMetric) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderBy{Expression: Distance(column, v, metric)})
	}
}

// quoteVectorColumn quotes a column name for PostgreSQL, leaving anything that
// is already qualified, quoted or an expression alone.
func quoteVectorColumn(column string) string {
	var trimmed = strings.TrimSpace(column)
	if trimmed == "" {
		return trimmed
	}
	if strings.ContainsAny(trimmed, `".( `) {
		return trimmed
	}
	return `"` + trimmed + `"`
}
