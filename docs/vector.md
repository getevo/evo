# Vector Columns

`types.Vector` stores an embedding — a list of floats — in a database column,
indexes it, and searches it by distance. It is what you build semantic search,
recommendations, deduplication or retrieval augmented generation on.

> ## PostgreSQL with pgvector only
>
> The vector column type comes from the [pgvector](https://github.com/pgvector/pgvector)
> extension. **Everything on this page needs PostgreSQL with pgvector installed.**
> There is no equivalent on the other drivers:
>
> | Driver | Column | Index | Distance search |
> |---|---|---|---|
> | **PostgreSQL + pgvector** | `vector(n)` | HNSW, IVFFlat | yes |
> | PostgreSQL without pgvector | migration fails | — | — |
> | MySQL / MariaDB | `JSON` | ignored | no |
> | SQLite | `TEXT` | ignored | no |
>
> On MySQL and SQLite a model carrying a vector still compiles and still
> migrates — the value is written as the same `[1,2,3]` text — but
> `vector_index` is ignored and `types.Nearest` produces PostgreSQL operators
> the database will reject. Treat a vector field as a PostgreSQL feature.

---

## Contents

- [Requirements](#requirements)
- [Declaring the column](#declaring-the-column)
- [What migration generates](#what-migration-generates)
- [Indexes](#indexes)
- [Writing and reading](#writing-and-reading)
- [Searching](#searching)
- [Distance metrics](#distance-metrics)
- [Complete example](#complete-example)
- [Pitfalls](#pitfalls)
- [API reference](#api-reference)
- [Testing](#testing)

---

## Requirements

pgvector has to be available to the server. The quickest way is the official
image:

```shell
docker run -d --name pgvector -p 127.0.0.1:5432:5432 \
  -e POSTGRES_USER=evo -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=app \
  pgvector/pgvector:pg17
```

On an existing server install the package (`apt install postgresql-17-pgvector`,
`brew install pgvector`, or build from source). Most managed providers — RDS,
Cloud SQL, Azure Database, Supabase, Neon — ship it and only need it enabled.

The extension itself is created by the migration:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

That statement needs the privilege to create an extension, which a restricted
application user may not have. Where that is the case, run it once by hand as a
superuser — `IF NOT EXISTS` then makes the migration's copy a no-op.

Use the PostgreSQL driver as usual:

```go
import "github.com/getevo/evo/v2/lib/pgsql"

func main() {
    evo.Setup(pgsql.Driver{})
    // ...
}
```

---

## Declaring the column

```go
import "github.com/getevo/evo/v2/lib/db/types"

type Document struct {
    ID        uint         `gorm:"column:id;primaryKey;autoIncrement"`
    Title     string       `gorm:"column:title;size:200"`
    Body      string       `gorm:"column:body;type:text"`
    Embedding types.Vector `gorm:"column:embedding;size:1536;vector_index:hnsw"`
}
```

Three things matter.

**The width.** A vector column has a fixed number of dimensions, and it is part
of the column type. Declare it with `size`, or spell the type out — both produce
`vector(1536)`:

```go
Embedding types.Vector `gorm:"size:1536"`            // from the size tag
Embedding types.Vector `gorm:"type:vector(1536)"`    // written out
Embedding types.Vector                               // vector, any width, cannot be indexed
```

The width has to match what your embedding model produces: 384 for
`all-MiniLM-L6-v2`, 768 for `nomic-embed-text`, 1536 for OpenAI
`text-embedding-3-small`, 1024 for Cohere `embed-english-v3.0`. A mismatch is
rejected by the database on write, not silently truncated.

**Nullability.** A vector column is nullable unless the model says otherwise.
That is deliberate: a row usually exists before anything has computed its
embedding, and a dimensioned vector column has no zero value to default to — an
empty vector is not a valid `vector(1536)`. `types.Vector.Value` writes an empty
or nil vector as `NULL`.

**The index.** Optional, and covered [below](#indexes). Without one, searches
still return the right answer by reading the whole table.

---

## What migration generates

For the model above (with `size:16` for brevity), `--migration-do` produces:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE IF NOT EXISTS "document"(
	"id" BIGSERIAL NOT NULL,
	"title" varchar(200) NOT NULL,
	"body" text NOT NULL,
	"embedding" vector(16) DEFAULT NULL NULL,
	PRIMARY KEY ("id")
);
COMMENT ON TABLE "document" IS '0.0.0';
CREATE INDEX "vx_document_embedding" ON "document" USING hnsw ("embedding" vector_cosine_ops);
```

The extension statement is emitted only when a vector column is actually being
created, so a schema that already matches its models stays quiet — the second
run of a migration produces nothing at all.

Adding a vector column to an existing table works the same way:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
ALTER TABLE "document" ADD COLUMN "embedding" vector(1536) DEFAULT NULL NULL;
CREATE INDEX "vx_document_embedding" ON "document" USING hnsw ("embedding" vector_cosine_ops);
```

---

## Indexes

Ask for an index with `vector_index`. The tag names the access method and the
metric:

| Tag | Index built |
|---|---|
| `vector_index` | HNSW, cosine distance |
| `vector_index:hnsw` | HNSW, cosine distance |
| `vector_index:hnsw,l2` | HNSW, straight line distance |
| `vector_index:hnsw,ip` | HNSW, inner product |
| `vector_index:hnsw,l1` | HNSW, taxicab distance (pgvector 0.7+) |
| `vector_index:ivfflat` | IVFFlat, cosine distance |
| `vector_index:ivfflat,l2` | IVFFlat, straight line distance |
| *(no tag)* | none — every search reads the whole table |

The index is named `vx_<table>_<column>`.

**HNSW or IVFFlat.** HNSW builds a graph: slower to build, more memory, best
recall, and it works on an empty table. IVFFlat clusters the rows: faster to
build, less memory, and it needs a representative amount of data present before
it is built to cluster sensibly. HNSW is the better default; reach for IVFFlat
when build time or memory is the constraint.

**The metric has to match the query.** An index carries an operator class, and
PostgreSQL only uses it for a query measuring distance the same way. A cosine
index with a cosine query:

```
Limit
  ->  Index Scan using vx_document_embedding on document
        Order By: (embedding <=> '[...]'::vector)
```

The same index with an L2 query — no index, the whole table sorted:

```
Limit
  ->  Sort
        Sort Key: ((embedding <-> '[...]'::vector))
        ->  Seq Scan on document
```

Both return correct rows. Only one of them scales. Check with `EXPLAIN` when it
matters.

**Changing the method or the metric** rebuilds the index on the next migration:

```sql
DROP INDEX IF EXISTS "vx_document_embedding";
CREATE INDEX "vx_document_embedding" ON "document" USING hnsw ("embedding" vector_l2_ops);
```

**A column with no declared width cannot be indexed** — pgvector refuses, since
it cannot build a graph over vectors of unknown size. The migration logs an
error naming the column and creates the table without the index rather than
failing.

**Build parameters are not exposed.** HNSW's `m` and `ef_construction`, and
IVFFlat's `lists`, are not settable through the tag. Where you need them, leave
`vector_index` off and create the index yourself:

```go
db.Exec(`CREATE INDEX IF NOT EXISTS vx_document_embedding ON document
         USING hnsw (embedding vector_cosine_ops) WITH (m = 24, ef_construction = 100)`)
```

Leaving the tag off also keeps the migrator from touching it: it only manages
the indexes the models declare.

---

## Writing and reading

A vector is an ordinary field. Assign it and save:

```go
document := Document{
    Title:     "Cancelling a subscription",
    Body:      "You can cancel at any time from the billing page.",
    Embedding: types.Vector{0.12, -0.04, 0.88 /* ... 1536 values ... */},
}
db.Create(&document)
```

Leave it unset and the column is `NULL`:

```go
db.Create(&Document{Title: "Draft", Body: "Not embedded yet."})
```

Fill it in later:

```go
db.Model(&document).Update("embedding", embed(document.Body))
```

Reading is symmetrical — `types.Vector` is a `[]float32`, so use it directly:

```go
var document Document
db.First(&document, 1)

fmt.Println(document.Embedding.Dimensions())  // 1536
fmt.Println(document.Embedding[0])            // 0.12
fmt.Println(document.Embedding.String())      // [0.12,-0.04,0.88,...]
```

The wrong width is an error from the database, which is what you want:

```
ERROR: expected 1536 dimensions, not 3 (SQLSTATE 22000)
```

---

## Searching

### The scope

`types.Nearest` orders rows by distance, closest first:

```go
var documents []Document
db.Model(&Document{}).
    Scopes(types.Nearest("embedding", queryVector, types.VectorCosine)).
    Limit(5).
    Find(&documents)
```

`ORDER BY` with `LIMIT` is the shape the index is built for. Always set a limit.

### The distance itself

`types.Distance` is the expression behind the scope. Select it to show how close
each hit is:

```go
var hits []struct {
    ID       uint
    Title    string
    Distance float64
}
db.Model(&Document{}).
    Select("id, title, ? AS distance", types.Distance("embedding", queryVector, types.VectorCosine)).
    Scopes(types.Nearest("embedding", queryVector, types.VectorCosine)).
    Limit(5).
    Find(&hits)
```

Filter on it to drop weak matches:

```go
db.Model(&Document{}).
    Where(gorm.Expr("? < ?", types.Distance("embedding", queryVector, types.VectorCosine), 0.35)).
    Scopes(types.Nearest("embedding", queryVector, types.VectorCosine)).
    Find(&documents)
```

A threshold is metric-specific: cosine distance runs 0 (identical direction) to 2
(opposite), so 0.3 is a reasonable cut for text. L2 has no upper bound and
depends on how your vectors are scaled.

Or pass it to `Clauses` where a scope does not fit:

```go
db.Model(&Document{}).
    Clauses(clause.OrderBy{Expression: types.Distance("embedding", queryVector, types.VectorCosine)}).
    Take(&document)
```

### Combining with ordinary conditions

Nothing about a vector search stops the rest of the query working:

```go
db.Model(&Document{}).
    Where("workspace_id = ? AND published = true", workspaceID).
    Scopes(types.Nearest("embedding", queryVector, types.VectorCosine)).
    Limit(10).
    Find(&documents)
```

Be aware of the trade-off: a filter that removes most rows can leave an
approximate index returning fewer than `LIMIT` results, because the index is
searched first and the filter applied after. For heavy filtering, either widen
the limit and trim in Go, or partition the data.

### Rows with no embedding

They have no distance, so they sort last. Exclude them when that matters:

```go
db.Model(&Document{}).
    Where("embedding IS NOT NULL").
    Scopes(types.Nearest("embedding", queryVector, types.VectorCosine)).
    Limit(5).
    Find(&documents)
```

### Qualified columns

A column name is quoted for you. Anything already qualified, quoted, or an
expression is passed through untouched, so joins work:

```go
types.Distance("embedding", v, types.VectorCosine)          // "embedding" <=> ?::vector
types.Distance(`"doc"."embedding"`, v, types.VectorCosine)  // "doc"."embedding" <=> ?::vector
```

---

## Distance metrics

| Constant | Operator | Operator class | Measures |
|---|---|---|---|
| `types.VectorCosine` | `<=>` | `vector_cosine_ops` | Angle only, magnitude ignored |
| `types.VectorL2` | `<->` | `vector_l2_ops` | Straight line distance |
| `types.VectorInnerProduct` | `<#>` | `vector_ip_ops` | Direction and magnitude, returned negated |
| `types.VectorL1` | `<+>` | `vector_l1_ops` | Taxicab distance (pgvector 0.7+) |

**Which one.** Cosine for text embeddings — it is what most models are trained
for, and it ignores document length. L2 when magnitude carries meaning, such as
coordinates or image features. Inner product when a larger magnitude should rank
higher, such as popularity baked into the vector; note pgvector returns it
negated, so the closest row is the most negative and ordering ascending still
works.

The three disagree, which is the point of choosing. For the query `[1,0,0]`
against `[1,0,0]`, `[0.8,0.6,0]`, `[0,1,0]`, `[-1,0,0]` and `[5,0.1,0]`:

| Metric | Order |
|---|---|
| cosine | `[1,0,0]`, `[5,0.1,0]`, `[0.8,0.6,0]`, `[0,1,0]`, `[-1,0,0]` |
| l2 | `[1,0,0]`, `[0.8,0.6,0]`, `[0,1,0]`, `[-1,0,0]`, `[5,0.1,0]` |
| ip | `[5,0.1,0]` first — it is the longest |

`[5,0.1,0]` points almost the same way as the query but is five times as long:
cosine calls it a near match, L2 calls it the furthest row of all, and inner
product calls it the best.

---

## Complete example

A runnable application is in [`examples/vector`](../examples/vector): semantic
search over a table of help articles, with a toy embedder so it runs with no API
key.

The model:

```go
const dimensions = 64

type Document struct {
    ID        uint         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    Title     string       `gorm:"column:title;size:200" json:"title"`
    Body      string       `gorm:"column:body;type:text" json:"body"`
    Embedding types.Vector `gorm:"column:embedding;size:64;vector_index:hnsw" json:"-"`
}

func (Document) TableName() string { return "document" }
```

The search handler:

```go
func search(request *evo.Request) any {
    var embedding = embed(request.Query("q").String())

    var hits []struct {
        ID       uint    `json:"id"`
        Title    string  `json:"title"`
        Body     string  `json:"body"`
        Distance float64 `json:"distance"`
    }
    db.Model(&Document{}).
        Select("id, title, body, ? AS distance", types.Distance("embedding", embedding, types.VectorCosine)).
        Scopes(types.Nearest("embedding", embedding, types.VectorCosine)).
        Limit(5).
        Find(&hits)

    return map[string]any{"query": request.Query("q").String(), "hits": hits}
}
```

Run it and ask a question. Write a `config.yml` pointing at your PostgreSQL in
the directory you run from — the header of `main.go` has one ready to copy — and
then:

```shell
cd examples/vector && go run .

curl -s '127.0.0.1:8080/search?q=how+do+i+cancel+a+subscription&k=2'
```

```json
{
  "success": true,
  "data": {
    "query": "how do i cancel a subscription",
    "hits": [
      { "id": 1, "title": "Cancelling a subscription", "distance": 0.600 },
      { "id": 2, "title": "Changing your plan",        "distance": 0.613 }
    ]
  }
}
```

```shell
curl -s '127.0.0.1:8080/search?q=my+card+was+declined&k=2'
```

```json
{ "hits": [ { "title": "Declined card", "distance": 0.528 },
            { "title": "Two factor authentication", "distance": 0.898 } ] }
```

Adding a document embeds it on the way in:

```shell
curl -s -X POST 127.0.0.1:8080/documents \
  -H 'Content-Type: application/json' \
  -d '{"title":"Refund window","body":"Refunds are available within 30 days."}'
```

The example's `embed` hashes words into buckets. It is enough to show the
database side, and its limits are the limits of keyword matching — ask it about
`2fa` and it will not find the article that says *two factor authentication*,
because the words never match. A real application calls an embedding model
instead and widens the column to that model's output.

---

## Pitfalls

**Changing the width breaks existing rows.** The migrator notices and emits the
ALTER, but PostgreSQL cannot convert stored vectors between widths:

```
-- column embedding type does not match. new:vector(64) old:vector(16)
ALTER TABLE "document" ALTER COLUMN "embedding" TYPE vector(64);
ERROR: expected 64 dimensions, not 16 (SQLSTATE 22000)
```

Changing embedding model means recomputing every embedding. Either clear the
column first (`UPDATE document SET embedding = NULL`), migrate, and backfill; or
add a second column, backfill it, and drop the old one.

**A metric mismatch silently costs you the index.** Correct rows, sequential
scan. See [Indexes](#indexes).

**Approximate indexes are approximate.** HNSW and IVFFlat trade recall for
speed: a search may miss a row that a full scan would have found. Where exactness
matters more than latency, leave the index off.

**Normalise if your model expects it.** Some models return unnormalised vectors.
Cosine handles that by construction; L2 and inner product do not.

**Do not put an embedding in an API response by accident.** 1536 floats is a lot
of JSON. Tag the field `json:"-"` unless a caller genuinely needs it.

---

## API reference

### `types.Vector`

```go
type Vector []float32
```

| Method | Purpose |
|---|---|
| `Scan(src any) error` | Reads `[1,2,3]` from the database, from `[]byte`, `string` or `nil` |
| `Value() (driver.Value, error)` | Writes `[1,2,3]`; an empty or nil vector becomes `NULL` |
| `String() string` | The stored form, `[1,2.5,-0.125]` |
| `Dimensions() int` | Length |
| `GormDataType() string` | `vector` |
| `GormDBDataType(db, field) string` | `vector(n)` on PostgreSQL, `JSON` on MySQL, `TEXT` on SQLite |

### `types.VectorMetric`

```go
const (
    VectorCosine       VectorMetric = "cosine"
    VectorL2           VectorMetric = "l2"
    VectorInnerProduct VectorMetric = "ip"
    VectorL1           VectorMetric = "l1"
)

func (m VectorMetric) OpsClass() string  // "vector_cosine_ops", ...
func (m VectorMetric) Valid() bool       // is it one of the four
```

### Query helpers

```go
// The distance between a column and a value, as an expression.
func types.Distance(column string, v Vector, metric VectorMetric) clause.Expr

// A scope ordering by that distance, closest first.
func types.Nearest(column string, v Vector, metric VectorMetric) func(*gorm.DB) *gorm.DB
```

### Extension helper

```go
// Creates the extension if it is missing. The migrator does this on its own
// when it creates a vector column; this is for code that builds its schema some
// other way, such as a test or a script.
func pgsql.EnsureVectorExtension(db *gorm.DB) error
```

---

## Testing

The vector tests need a PostgreSQL with pgvector and skip without one:

```shell
docker run -d --name evo-pgvector-test -p 127.0.0.1:5459:5432 \
  -e POSTGRES_USER=evo -e POSTGRES_PASSWORD=evotest -e POSTGRES_DB=evo_vector \
  pgvector/pgvector:pg17

EVO_PG_DSN="host=127.0.0.1 port=5459 user=evo password=evotest dbname=evo_vector sslmode=disable" \
  go test -run Vector ./lib/pgsql/ ./lib/db/types/
```

They cover the column type from both tag forms, nullability, the index and its
operator class, a metric change rebuilding it, distance search under three
metrics, rows without embeddings, wrong widths being rejected, and — the one that
matters most — that a second migration of an unchanged schema produces nothing.

---

## See also

- [PostgreSQL Driver](pgsql.md)
- [Database](database.md)
- [Database Migration](migration.md)
- [pgvector](https://github.com/pgvector/pgvector)
