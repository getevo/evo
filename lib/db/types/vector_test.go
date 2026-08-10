package types

import (
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestVectorScanAndValue(t *testing.T) {
	var cases = []struct {
		name string
		src  any
		want Vector
	}{
		{"pgvector bytes", []byte("[1,2,3]"), Vector{1, 2, 3}},
		{"pgvector string", "[1,2,3]", Vector{1, 2, 3}},
		{"json array", []byte("[0.5,-0.25]"), Vector{0.5, -0.25}},
		{"spaces", "[ 1 , 2 ]", Vector{1, 2}},
		{"scientific", "[1e-3,2.5e2]", Vector{0.001, 250}},
		{"empty vector", "[]", Vector{}},
		{"nil", nil, nil},
		{"null literal", "null", nil},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			var v Vector
			if err := v.Scan(item.src); err != nil {
				t.Fatalf("scanning %v failed: %v", item.src, err)
			}
			if len(v) != len(item.want) {
				t.Fatalf("scanned %v, want %v", v, item.want)
			}
			for i := range v {
				if v[i] != item.want[i] {
					t.Fatalf("scanned %v, want %v", v, item.want)
				}
			}
		})
	}
}

func TestVectorScanRejectsMalformedInput(t *testing.T) {
	for _, src := range []any{"1,2,3", "[1,2", "[a,b]", "[1,,2]", 42} {
		var v Vector
		if err := v.Scan(src); err == nil {
			t.Errorf("scanning %#v was accepted", src)
		}
	}
}

// The stored form has to be one both pgvector and a JSON column accept, so it
// is rendered without spaces and without trailing zeros.
func TestVectorValue(t *testing.T) {
	value, err := Vector{1, 2.5, -0.125}.Value()
	if err != nil {
		t.Fatal(err)
	}
	if value != "[1,2.5,-0.125]" {
		t.Fatalf("stored as %v, want [1,2.5,-0.125]", value)
	}
}

// A dimensioned vector column rejects a zero length value, so "no embedding
// yet" has to reach the database as NULL.
func TestVectorValueOfEmptyIsNull(t *testing.T) {
	for _, v := range []Vector{nil, {}} {
		value, err := v.Value()
		if err != nil {
			t.Fatal(err)
		}
		if value != nil {
			t.Fatalf("empty vector stored as %v, want NULL", value)
		}
	}
}

func TestVectorRoundTrip(t *testing.T) {
	var original = Vector{0.1, -0.2, 3.5, 0}
	value, err := original.Value()
	if err != nil {
		t.Fatal(err)
	}
	var back Vector
	if err := back.Scan(value.(string)); err != nil {
		t.Fatal(err)
	}
	if back.String() != original.String() {
		t.Fatalf("round tripped %s into %s", original, back)
	}
}

func TestVectorGormDBDataType(t *testing.T) {
	var cases = []struct {
		dialect string
		size    int
		want    string
	}{
		{"postgres", 1536, "vector(1536)"},
		{"postgres", 0, "vector"},
		{"mysql", 1536, "JSON"},
		{"sqlite", 1536, "TEXT"},
		{"sqlserver", 1536, "TEXT"},
	}

	for _, item := range cases {
		var db = &gorm.DB{Config: &gorm.Config{Dialector: fakeDialector{name: item.dialect}}}
		var got = Vector{}.GormDBDataType(db, &schema.Field{Size: item.size})
		if got != item.want {
			t.Errorf("%s with size %d gave %q, want %q", item.dialect, item.size, got, item.want)
		}
	}
}

func TestVectorMetrics(t *testing.T) {
	var cases = []struct {
		metric   VectorMetric
		operator string
		ops      string
	}{
		{VectorCosine, "<=>", "vector_cosine_ops"},
		{VectorL2, "<->", "vector_l2_ops"},
		{VectorInnerProduct, "<#>", "vector_ip_ops"},
		{VectorL1, "<+>", "vector_l1_ops"},
	}

	for _, item := range cases {
		if got := item.metric.operator(); got != item.operator {
			t.Errorf("%s uses operator %q, want %q", item.metric, got, item.operator)
		}
		if got := item.metric.OpsClass(); got != item.ops {
			t.Errorf("%s uses operator class %q, want %q", item.metric, got, item.ops)
		}
		if !item.metric.Valid() {
			t.Errorf("%s is not reported as valid", item.metric)
		}
	}
	if VectorMetric("euclidean").Valid() {
		t.Error("an unknown metric is reported as valid")
	}
}

// PostgreSQL cannot infer the type of a bare placeholder next to a distance
// operator, so the parameter carries an explicit cast.
func TestDistanceExpression(t *testing.T) {
	var expr = Distance("embedding", Vector{1, 0, 0}, VectorCosine)
	if expr.SQL != `"embedding" <=> ?::vector` {
		t.Fatalf("expression is %q", expr.SQL)
	}
	if len(expr.Vars) != 1 || expr.Vars[0] != "[1,0,0]" {
		t.Fatalf("parameters are %v", expr.Vars)
	}
}

// An expression or an already qualified name is passed through untouched, so a
// join can say which table it means.
func TestDistanceLeavesQualifiedColumnsAlone(t *testing.T) {
	for _, column := range []string{`"doc"."embedding"`, "doc.embedding", `"embedding"`} {
		var expr = Distance(column, Vector{1}, VectorL2)
		if expr.SQL != column+" <-> ?::vector" {
			t.Errorf("column %q became %q", column, expr.SQL)
		}
	}
}

// fakeDialector reports a name and nothing else, which is all
// GormDBDataType looks at.
type fakeDialector struct {
	postgres.Dialector
	name string
}

func (d fakeDialector) Name() string { return d.name }
