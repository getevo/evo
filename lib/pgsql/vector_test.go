package pgsql

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/db/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// These tests need a PostgreSQL with the pgvector extension available:
//
//	docker run -d --name evo-pgvector-test -p 127.0.0.1:5459:5432 \
//	  -e POSTGRES_USER=evo -e POSTGRES_PASSWORD=evotest -e POSTGRES_DB=evo_vector \
//	  pgvector/pgvector:pg17
//
//	EVO_PG_DSN="host=127.0.0.1 port=5459 user=evo password=evotest dbname=evo_vector sslmode=disable" \
//	  go test -run Vector ./lib/pgsql/
//
// Without EVO_PG_DSN they skip.

const vectorTable = "evo_vector_doc"

// vectorDoc declares the width with the size tag and asks for an HNSW index
// measuring cosine distance.
type vectorDoc struct {
	ID        uint         `gorm:"column:id;primaryKey;autoIncrement"`
	Title     string       `gorm:"column:title;size:64"`
	Embedding types.Vector `gorm:"column:embedding;size:3;vector_index:hnsw"`
}

func (vectorDoc) TableName() string { return vectorTable }

// vectorDocL2 is the same table asking for straight line distance instead,
// used to check that changing the metric rebuilds the index.
type vectorDocL2 struct {
	ID        uint         `gorm:"column:id;primaryKey;autoIncrement"`
	Title     string       `gorm:"column:title;size:64"`
	Embedding types.Vector `gorm:"column:embedding;size:3;vector_index:hnsw,l2"`
}

func (vectorDocL2) TableName() string { return vectorTable }

// vectorDocTyped spells the column type out in full instead of using size.
type vectorDocTyped struct {
	ID        uint         `gorm:"column:id;primaryKey;autoIncrement"`
	Embedding types.Vector `gorm:"column:embedding;type:vector(3)"`
}

func (vectorDocTyped) TableName() string { return "evo_vector_typed" }

// vectorDocBare declares no width at all, which pgvector allows but cannot
// index.
type vectorDocBare struct {
	ID        uint         `gorm:"column:id;primaryKey;autoIncrement"`
	Embedding types.Vector `gorm:"column:embedding;vector_index:hnsw"`
}

func (vectorDocBare) TableName() string { return "evo_vector_bare" }

func openVectorDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("EVO_PG_DSN")
	if dsn == "" {
		t.Skip("EVO_PG_DSN is not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("cannot reach the database: %v", err)
	}
	schema.SetDialect(NewPGDialect("public"))
	return db
}

// migrate runs the generated script the way DoMigration does, and returns the
// statements it executed.
func migrate(t *testing.T, db *gorm.DB, models ...any) []string {
	t.Helper()
	schema.ResetMigrations()
	schema.UseModel(db, models...)

	var script = schema.GetMigrationScript(db)
	var executed []string
	for _, query := range script {
		query = strings.TrimSpace(query)
		if query == "" || strings.HasPrefix(query, "--") {
			continue
		}
		if err := db.Exec(query).Error; err != nil {
			t.Fatalf("statement failed: %s\n%v", query, err)
		}
		executed = append(executed, query)
	}
	return executed
}

// pendingStatements returns the executable statements the migrator would run
// now, ignoring its comments.
func pendingStatements(t *testing.T, db *gorm.DB, models ...any) []string {
	t.Helper()
	schema.ResetMigrations()
	schema.UseModel(db, models...)

	var pending []string
	for _, query := range schema.GetMigrationScript(db) {
		query = strings.TrimSpace(query)
		if query == "" || strings.HasPrefix(query, "--") {
			continue
		}
		pending = append(pending, query)
	}
	return pending
}

func TestVectorMigration(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "` + vectorTable + `"`)

	var executed = migrate(t, db, vectorDoc{})
	var script = strings.Join(executed, "\n")

	// The type has to be made available before a column can use it.
	if !strings.Contains(script, "CREATE EXTENSION IF NOT EXISTS vector") {
		t.Errorf("the extension was not created:\n%s", script)
	}
	// The width comes from the size tag.
	if !strings.Contains(script, "vector(3)") {
		t.Errorf("the column was not declared as vector(3):\n%s", script)
	}
	// An embedding is usually computed after the row exists.
	if strings.Contains(script, `"embedding" vector(3) NOT NULL`) {
		t.Errorf("the column was declared NOT NULL:\n%s", script)
	}
	// The index needs both an access method and an operator class.
	if !strings.Contains(script, `USING hnsw ("embedding" vector_cosine_ops)`) {
		t.Errorf("the vector index was not created:\n%s", script)
	}

	// What the database actually holds.
	var columnType string
	var nullable string
	if err := db.Raw(`
		SELECT format_type(a.atttypid, a.atttypmod), CASE WHEN a.attnotnull THEN 'NO' ELSE 'YES' END
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		WHERE c.relname = ? AND a.attname = 'embedding'
	`, vectorTable).Row().Scan(&columnType, &nullable); err != nil {
		t.Fatal(err)
	}
	if columnType != "vector(3)" {
		t.Errorf("the column is %s, want vector(3)", columnType)
	}
	if nullable != "YES" {
		t.Error("the column is NOT NULL")
	}

	var method, definition string
	if err := db.Raw(`
		SELECT am.amname, pg_get_indexdef(i.oid)
		FROM pg_class i
		JOIN pg_index ix ON ix.indexrelid = i.oid
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_am am ON am.oid = i.relam
		WHERE t.relname = ? AND i.relname LIKE 'vx_%'
	`, vectorTable).Row().Scan(&method, &definition); err != nil {
		t.Fatalf("no vector index was created: %v", err)
	}
	if method != "hnsw" {
		t.Errorf("the index uses %s, want hnsw", method)
	}
	if !strings.Contains(definition, "vector_cosine_ops") {
		t.Errorf("the index is %s, want cosine", definition)
	}
}

// The check that catches every introspection mistake at once: a schema that
// matches the model has nothing to migrate. If the width, the nullability or
// the index were read back wrongly, the migrator would rewrite the column on
// every boot.
func TestVectorMigrationIsIdempotent(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "` + vectorTable + `"`)
	migrate(t, db, vectorDoc{})

	if pending := pendingStatements(t, db, vectorDoc{}); len(pending) > 0 {
		t.Fatalf("a second migration would run %d statements:\n%s",
			len(pending), strings.Join(pending, "\n"))
	}

	// Not one comment either. A script that holds nothing but comments still
	// counts as a migration to DoMigration, which would then record a new hash
	// on every boot.
	schema.ResetMigrations()
	schema.UseModel(db, vectorDoc{})
	if script := schema.GetMigrationScript(db); len(script) > 0 {
		t.Fatalf("the second migration script is not empty:\n%s", strings.Join(script, "\n"))
	}
}

// A vector index is defined by its access method and operator class as much as
// by its column, so changing the metric has to rebuild it.
func TestVectorIndexMetricChangeRebuildsTheIndex(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "` + vectorTable + `"`)
	migrate(t, db, vectorDoc{})

	var pending = pendingStatements(t, db, vectorDocL2{})
	var script = strings.Join(pending, "\n")
	if !strings.Contains(script, "DROP INDEX") || !strings.Contains(script, "vector_l2_ops") {
		t.Fatalf("changing the metric did not rebuild the index:\n%s", script)
	}

	migrate(t, db, vectorDocL2{})
	if pending := pendingStatements(t, db, vectorDocL2{}); len(pending) > 0 {
		t.Fatalf("the rebuilt index still reports a pending migration:\n%s",
			strings.Join(pending, "\n"))
	}
}

// The width can also be spelled out in the type tag, which reaches the
// migrator by a different route than size.
func TestVectorTypeTag(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "evo_vector_typed"`)

	var script = strings.Join(migrate(t, db, vectorDocTyped{}), "\n")
	if !strings.Contains(script, "vector(3)") {
		t.Errorf("the type tag was not honoured:\n%s", script)
	}
	if pending := pendingStatements(t, db, vectorDocTyped{}); len(pending) > 0 {
		t.Errorf("a second migration would run:\n%s", strings.Join(pending, "\n"))
	}
}

// A column with no declared width is legal and usable, but pgvector cannot
// index it. The table still has to be created.
func TestVectorWithoutDimensionsIsNotIndexed(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "evo_vector_bare"`)

	var script = strings.Join(migrate(t, db, vectorDocBare{}), "\n")
	if !strings.Contains(script, `"embedding" vector`) {
		t.Errorf("the column was not created:\n%s", script)
	}
	if strings.Contains(script, "USING hnsw") {
		t.Errorf("an index was attempted on a column with no width:\n%s", script)
	}
	if pending := pendingStatements(t, db, vectorDocBare{}); len(pending) > 0 {
		t.Errorf("a second migration would run:\n%s", strings.Join(pending, "\n"))
	}
}

func TestVectorSearch(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "` + vectorTable + `"`)
	migrate(t, db, vectorDoc{})

	var rows = []vectorDoc{
		{Title: "x axis", Embedding: types.Vector{1, 0, 0}},
		{Title: "y axis", Embedding: types.Vector{0, 1, 0}},
		{Title: "between", Embedding: types.Vector{0.8, 0.6, 0}},
		{Title: "far", Embedding: types.Vector{-1, 0, 0}},
		// Points almost the same way as the query but is five times as long:
		// cosine calls it a near match, straight line distance calls it the
		// furthest row of all. It is here so that the two metrics disagree —
		// otherwise this test would pass with the wrong operator.
		{Title: "long", Embedding: types.Vector{5, 0.1, 0}},
		{Title: "no embedding yet"}, // the case the nullable column exists for
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("inserting failed: %v", err)
	}

	var query = types.Vector{1, 0, 0}

	// Reading a vector back has to produce what was written.
	var stored vectorDoc
	if err := db.Where("title = ?", "between").Take(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Embedding.String() != "[0.8,0.6,0]" {
		t.Errorf("read back %s, want [0.8,0.6,0]", stored.Embedding)
	}

	// The two metrics rank the same rows differently, which is the point: an
	// order that matched both would not prove the operator reached the database.
	var cases = []struct {
		metric types.VectorMetric
		order  []string
	}{
		{types.VectorCosine, []string{"x axis", "long", "between", "y axis", "far"}},
		{types.VectorL2, []string{"x axis", "between", "y axis", "far", "long"}},
	}

	for _, item := range cases {
		var found []vectorDoc
		if err := db.Model(&vectorDoc{}).
			Scopes(types.Nearest("embedding", query, item.metric)).
			Find(&found).Error; err != nil {
			t.Fatalf("%s search failed: %v", item.metric, err)
		}
		if len(found) != len(rows) {
			t.Fatalf("%s search returned %d rows, want %d", item.metric, len(found), len(rows))
		}

		var titles []string
		for _, doc := range found {
			titles = append(titles, doc.Title)
		}
		for i, title := range item.order {
			if titles[i] != title {
				t.Fatalf("%s order is %v, want %v then the row with no embedding",
					item.metric, titles, item.order)
			}
		}
		// A row with no embedding has no distance, so it sorts last.
		if titles[len(titles)-1] != "no embedding yet" {
			t.Errorf("%s order is %v: the row with no embedding is not last", item.metric, titles)
		}
	}

	// The distance itself, selected alongside the row.
	var scored struct {
		Title    string
		Distance float64
	}
	if err := db.Model(&vectorDoc{}).
		Select("title, ? AS distance", types.Distance("embedding", query, types.VectorCosine)).
		Where("title = ?", "x axis").
		Take(&scored).Error; err != nil {
		t.Fatalf("selecting the distance failed: %v", err)
	}
	if scored.Distance > 1e-6 {
		t.Errorf("the distance to an identical vector is %v, want 0", scored.Distance)
	}

	// Inner product ranks magnitude as well as direction, and pgvector returns
	// it negated, so the closest row is the most negative one — the long vector
	// rather than the identical one. A third ranking from the same rows.
	var best vectorDoc
	if err := db.Model(&vectorDoc{}).
		Clauses(clause.OrderBy{Expression: types.Distance("embedding", types.Vector{2, 0, 0}, types.VectorInnerProduct)}).
		Take(&best).Error; err != nil {
		t.Fatalf("inner product search failed: %v", err)
	}
	if best.Title != "long" {
		t.Errorf("inner product ranked %q first, want long", best.Title)
	}
}

func TestVectorRejectsWrongDimensions(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "` + vectorTable + `"`)
	migrate(t, db, vectorDoc{})

	var row = vectorDoc{Title: "too wide", Embedding: types.Vector{1, 2, 3, 4}}
	if err := db.Create(&row).Error; err == nil {
		t.Fatal("a four dimensional vector was accepted by a vector(3) column")
	}
}

func TestIsVectorTypeAndDimensions(t *testing.T) {
	var cases = []struct {
		in     string
		vector bool
		dims   int
	}{
		{"vector", true, 0},
		{"vector(3)", true, 3},
		{"VECTOR(1536)", true, 1536},
		{" vector(768) ", true, 768},
		{"vector()", true, 0},
		{"varchar(255)", false, 0},
		{"text", false, 0},
		{"", false, 0},
	}

	for _, item := range cases {
		if got := isVectorType(item.in); got != item.vector {
			t.Errorf("isVectorType(%q) = %v", item.in, got)
		}
		if got := vectorDimensions(item.in); got != item.dims {
			t.Errorf("vectorDimensions(%q) = %d, want %d", item.in, got, item.dims)
		}
	}
}

func TestVectorIndexTagParsing(t *testing.T) {
	var column = schema.Column{Name: "embedding", Type: "vector(3)"}

	var cases = []struct {
		tag    string
		hasTag bool
		method string
		ops    string
		nilOut bool
	}{
		{"", false, "", "", true}, // no tag, no index
		{"", true, "hnsw", "vector_cosine_ops", false},
		{"hnsw", true, "hnsw", "vector_cosine_ops", false},
		{"ivfflat", true, "ivfflat", "vector_cosine_ops", false},
		{"hnsw,l2", true, "hnsw", "vector_l2_ops", false},
		{"ivfflat,ip", true, "ivfflat", "vector_ip_ops", false},
		{"HNSW,L2", true, "hnsw", "vector_l2_ops", false},
		{"btree", true, "", "", true},          // not a vector method
		{"hnsw,haversine", true, "", "", true}, // not a metric
	}

	for _, item := range cases {
		var index = vectorIndex("doc", column, item.tag, item.hasTag)
		if item.nilOut {
			if index != nil {
				t.Errorf("tag %q built %+v, want no index", item.tag, index)
			}
			continue
		}
		if index == nil {
			t.Errorf("tag %q built no index", item.tag)
			continue
		}
		if index.Method != item.method || index.Ops != item.ops {
			t.Errorf("tag %q built %s/%s, want %s/%s", item.tag, index.Method, index.Ops, item.method, item.ops)
		}
		if index.Name != "vx_doc_embedding" {
			t.Errorf("tag %q named the index %s", item.tag, index.Name)
		}
	}

	// A column with no declared width cannot be indexed.
	if index := vectorIndex("doc", schema.Column{Name: "embedding", Type: "vector"}, "hnsw", true); index != nil {
		t.Error("a column with no width was given an index")
	}
}

// plainDoc has no vector column. It is here because the vector work changed
// the introspection queries every table goes through, and a schema that
// matches its model must still have nothing to migrate.
type plainDoc struct {
	ID        uint       `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string     `gorm:"column:name;size:120;index"`
	Email     string     `gorm:"column:email;size:190;uniqueIndex"`
	Score     float64    `gorm:"column:score"`
	Active    bool       `gorm:"column:active"`
	Notes     *string    `gorm:"column:notes;type:text"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (plainDoc) TableName() string { return "evo_plain_doc" }

func TestPlainTableMigrationIsStillIdempotent(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "evo_plain_doc"`)

	migrate(t, db, plainDoc{})
	if pending := pendingStatements(t, db, plainDoc{}); len(pending) > 0 {
		t.Fatalf("a second migration of a table with no vector would run %d statements:\n%s",
			len(pending), strings.Join(pending, "\n"))
	}
}

// The indexes of an ordinary table have to be read back with their columns
// intact, which the index query now also reports a method for.
func TestPlainTableIndexesAreReadBack(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "evo_plain_doc"`)
	migrate(t, db, plainDoc{})

	p := NewPGDialect("public")
	database := p.GetCurrentDatabase(db)
	tables := p.introspectTables(db, database)
	p.assembleColumns(p.introspectColumns(db, database), tables)
	p.introspectIndexes(db, database, tables)

	tbl := tables.GetTable("evo_plain_doc")
	if tbl == nil {
		t.Fatal("the table was not introspected")
	}
	if len(tbl.Indexes) == 0 {
		t.Fatal("no indexes were introspected")
	}
	for _, index := range tbl.Indexes {
		if len(index.Columns) == 0 {
			t.Errorf("index %s came back with no columns", index.Name)
		}
		if index.Method != "btree" {
			t.Errorf("index %s came back with method %q, want btree", index.Name, index.Method)
		}
		if !strings.Contains(index.Definition, "evo_plain_doc") {
			t.Errorf("index %s came back with definition %q", index.Name, index.Definition)
		}
	}
}

// Filtering on the distance rather than only ordering by it. The expression is
// nested inside another one, which is what the documentation shows.
func TestVectorDistanceFilter(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "` + vectorTable + `"`)
	migrate(t, db, vectorDoc{})

	if err := db.Create(&[]vectorDoc{
		{Title: "same", Embedding: types.Vector{1, 0, 0}},
		{Title: "near", Embedding: types.Vector{0.8, 0.6, 0}},
		{Title: "opposite", Embedding: types.Vector{-1, 0, 0}},
	}).Error; err != nil {
		t.Fatal(err)
	}

	var query = types.Vector{1, 0, 0}
	var matches []vectorDoc
	if err := db.Model(&vectorDoc{}).
		Where(gorm.Expr("? < ?", types.Distance("embedding", query, types.VectorCosine), 0.25)).
		Order("title").
		Find(&matches).Error; err != nil {
		t.Fatalf("filtering on distance failed: %v", err)
	}

	if len(matches) != 2 {
		var titles []string
		for _, doc := range matches {
			titles = append(titles, doc.Title)
		}
		t.Fatalf("the filter matched %v, want same and near", titles)
	}
	if matches[0].Title != "near" || matches[1].Title != "same" {
		t.Fatalf("the filter matched %s and %s", matches[0].Title, matches[1].Title)
	}
}

// plainDocLater is plainDoc with an embedding added afterwards, which is the
// ALTER path rather than the CREATE path: the extension has to be ensured there
// too, and the index built after the column exists.
type plainDocLater struct {
	ID        uint         `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string       `gorm:"column:name;size:120;index"`
	Email     string       `gorm:"column:email;size:190;uniqueIndex"`
	Score     float64      `gorm:"column:score"`
	Active    bool         `gorm:"column:active"`
	Notes     *string      `gorm:"column:notes;type:text"`
	CreatedAt time.Time    `gorm:"column:created_at"`
	DeletedAt *time.Time   `gorm:"column:deleted_at"`
	Embedding types.Vector `gorm:"column:embedding;size:4;vector_index:ivfflat,l2"`
}

func (plainDocLater) TableName() string { return "evo_plain_doc" }

func TestVectorColumnAddedToAnExistingTable(t *testing.T) {
	db := openVectorDB(t)
	db.Exec(`DROP TABLE IF EXISTS "evo_plain_doc"`)

	// A table with rows and no vector column.
	migrate(t, db, plainDoc{})
	if err := db.Exec(`INSERT INTO "evo_plain_doc" ("name","email","score","active") VALUES ('a','a@example.com',1,true)`).Error; err != nil {
		t.Fatal(err)
	}

	var script = strings.Join(migrate(t, db, plainDocLater{}), "\n")

	if !strings.Contains(script, "CREATE EXTENSION IF NOT EXISTS vector") {
		t.Errorf("the extension was not ensured on the ALTER path:\n%s", script)
	}
	if !strings.Contains(script, `ADD COLUMN "embedding" vector(4)`) {
		t.Errorf("the column was not added:\n%s", script)
	}
	if !strings.Contains(script, `USING ivfflat ("embedding" vector_l2_ops)`) {
		t.Errorf("the index was not created:\n%s", script)
	}
	// The extension has to come before the column that needs it.
	if strings.Index(script, "CREATE EXTENSION") > strings.Index(script, "ADD COLUMN") {
		t.Errorf("the extension was ensured after the column:\n%s", script)
	}

	// The existing row keeps its data and gets a NULL embedding.
	var count int64
	db.Table("evo_plain_doc").Where("embedding IS NULL").Count(&count)
	if count != 1 {
		t.Errorf("the existing row has %d null embeddings, want 1", count)
	}

	if pending := pendingStatements(t, db, plainDocLater{}); len(pending) > 0 {
		t.Fatalf("a second migration would run:\n%s", strings.Join(pending, "\n"))
	}
}
