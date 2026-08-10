// Example application doing semantic search over its own table with pgvector.
//
// It needs PostgreSQL with the pgvector extension available — the vector column
// type exists nowhere else. The image below has it:
//
//	docker run -d --name evo-pgvector -p 127.0.0.1:5459:5432 \
//	  -e POSTGRES_USER=evo -e POSTGRES_PASSWORD=evotest -e POSTGRES_DB=evo_vector \
//	  pgvector/pgvector:pg17
//
// The configuration is read from the working directory, so write this
// config.yml in the directory you run from — the repository ignores *.yml, so it
// is not shipped:
//
//	Database:
//	  Enabled: true
//	  Type: postgres
//	  Server: 127.0.0.1:5459
//	  Username: evo
//	  Password: evotest
//	  Database: evo_vector
//	  SSLMode: "false"
//	HTTP:
//	  Host: 127.0.0.1
//	  Port: 8080
//
// Then:
//
//	cd examples/vector && go run .
//
//	curl -s '127.0.0.1:8080/search?q=how+do+i+cancel+a+subscription'
//	curl -s '127.0.0.1:8080/search?q=payment+card+declined&metric=l2'
//	curl -s -X POST 127.0.0.1:8080/documents \
//	  -H 'Content-Type: application/json' \
//	  -d '{"title":"Refund window","body":"Refunds are available within 30 days of purchase."}'
package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"strings"

	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/application"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/db/types"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/outcome"
	"github.com/getevo/evo/v2/lib/pgsql"
	"gorm.io/gorm/clause"
)

// dimensions is small for an embedding, but wide enough that the toy embedder
// below does not collide every other word. A real
// embedding is 384, 768 or 1536 wide depending on the model that produced it.
const dimensions = 64

// Document is the searchable table.
//
//   - size declares the width of the column, which is part of its type
//   - vector_index asks for an HNSW index measuring cosine distance, which is
//     the usual pairing for text embeddings
//
// The embedding column is nullable: a document can exist before anything has
// computed its vector.
type Document struct {
	ID        uint         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title     string       `gorm:"column:title;size:200" json:"title"`
	Body      string       `gorm:"column:body;type:text" json:"body"`
	Embedding types.Vector `gorm:"column:embedding;size:64;vector_index:hnsw" json:"-"`
}

func (Document) TableName() string { return "document" }

// --- the application -------------------------------------------------------

type App struct{}

func (App) Name() string                   { return "vector" }
func (App) Priority() application.Priority { return application.DEFAULT }

// WhenReady runs one search at boot using the expression form directly, to show
// what it looks like outside an HTTP handler.
func (App) WhenReady() error {
	var closest Document
	if err := db.Model(&Document{}).
		Where("embedding IS NOT NULL").
		Clauses(clause.OrderBy{Expression: types.Distance("embedding", embed("billing invoice"), types.VectorCosine)}).
		Take(&closest).Error; err != nil {
		return err
	}
	log.Info(`closest document to "billing invoice": ` + closest.Title)
	return nil
}

func (App) Register() error {
	db.UseModel(Document{})

	// Creates the extension, the vector(64) column and the HNSW index.
	if err := db.DoMigration(); err != nil {
		return err
	}

	return seed()
}

func (App) Router() error {
	evo.Get("/search", search)
	evo.Post("/documents", create)
	return nil
}

// search embeds the query and returns the closest documents, with the distance
// that ranked them.
//
//	/search?q=cancel+my+plan&k=3&metric=cosine
func search(request *evo.Request) any {
	var query = strings.TrimSpace(request.Query("q").String())
	if query == "" {
		return outcome.Json(map[string]string{"error": "q is required"}).Status(400)
	}

	var metric = types.VectorCosine
	if v := request.Query("metric").String(); v != "" {
		metric = types.VectorMetric(v)
		if !metric.Valid() {
			return outcome.Json(map[string]string{
				"error": "metric must be cosine, l2, ip or l1",
			}).Status(400)
		}
	}

	var limit = request.Query("k").Int()
	if limit <= 0 || limit > 50 {
		limit = 5
	}

	var embedding = embed(query)

	// The distance is selected alongside the row so the caller can see how
	// close each hit is, and Nearest puts the closest first.
	var hits []struct {
		ID       uint    `json:"id"`
		Title    string  `json:"title"`
		Body     string  `json:"body"`
		Distance float64 `json:"distance"`
	}
	if err := db.Model(&Document{}).
		Select("id, title, body, ? AS distance", types.Distance("embedding", embedding, metric)).
		Scopes(types.Nearest("embedding", embedding, metric)).
		Limit(limit).
		Find(&hits).Error; err != nil {
		log.Error("search failed", "error", err)
		return outcome.Json(map[string]string{"error": err.Error()}).Status(500)
	}

	return map[string]any{
		"query":  query,
		"metric": metric,
		"hits":   hits,
	}
}

// create stores a document and its embedding.
func create(request *evo.Request) any {
	var input struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := request.BodyParser(&input); err != nil {
		return outcome.Json(map[string]string{"error": err.Error()}).Status(400)
	}
	if strings.TrimSpace(input.Body) == "" {
		return outcome.Json(map[string]string{"error": "body is required"}).Status(400)
	}

	var document = Document{
		Title:     input.Title,
		Body:      input.Body,
		Embedding: embed(input.Title + " " + input.Body),
	}
	if err := db.Create(&document).Error; err != nil {
		return outcome.Json(map[string]string{"error": err.Error()}).Status(500)
	}
	return document
}

// --- embeddings ------------------------------------------------------------

// embed turns text into a vector. This one hashes words into buckets and
// normalises the result, which is enough to demonstrate the database side
// without an API key: documents sharing words end up close together.
//
// A real application replaces this with a call to an embedding model and widens
// the column to match — the size tag, the vector(n) column and the model's
// output all have to agree.
func embed(text string) types.Vector {
	var vector = make(types.Vector, dimensions)
	for _, word := range strings.Fields(strings.ToLower(text)) {
		word = strings.Trim(word, ".,!?;:'\"()")
		if word == "" {
			continue
		}
		var h = fnv.New32a()
		_, _ = h.Write([]byte(word))
		vector[int(h.Sum32())%dimensions] += 1
	}

	// Cosine distance ignores magnitude, but normalising keeps the numbers
	// small and makes l2 behave sensibly too.
	var sum float64
	for _, v := range vector {
		sum += float64(v) * float64(v)
	}
	if sum == 0 {
		return vector
	}
	var norm = float32(math.Sqrt(sum))
	for i := range vector {
		vector[i] /= norm
	}
	return vector
}

// --- seed data -------------------------------------------------------------

func seed() error {
	var count int64
	if err := db.Model(&Document{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var documents = []Document{
		{Title: "Cancelling a subscription", Body: "You can cancel your subscription at any time from the billing page. Access continues until the end of the paid period."},
		{Title: "Changing your plan", Body: "Upgrading a plan takes effect immediately. Downgrading a plan takes effect at the next renewal."},
		{Title: "Declined card", Body: "A declined card is usually a bank block or an expired card. Update the card on the billing page and retry the payment."},
		{Title: "Invoices and receipts", Body: "Every payment produces an invoice. Invoices are available as PDF from the billing page."},
		{Title: "Resetting a password", Body: "Use the forgotten password link on the sign in page. The reset link is valid for one hour."},
		{Title: "Two factor authentication", Body: "Two factor authentication can be enabled in security settings using any authenticator app."},
		{Title: "Exporting your data", Body: "A full export of your account data can be requested from settings and arrives by email as a zip archive."},
		// A document with no embedding yet: the column is nullable, and rows
		// like this sort last in every search because they have no distance.
		{Title: "Draft article", Body: "This one has not been embedded yet."},
	}
	for i := range documents {
		if documents[i].Title == "Draft article" {
			continue
		}
		documents[i].Embedding = embed(documents[i].Title + " " + documents[i].Body)
	}

	if err := db.Create(&documents).Error; err != nil {
		return err
	}
	log.Info(fmt.Sprintf("seeded %d documents", len(documents)))
	return nil
}

func main() {
	if err := evo.Setup(pgsql.Driver{}); err != nil {
		fmt.Println("setup failed:", err)
		return
	}

	evo.Register(App{})

	if err := evo.Run(); err != nil {
		fmt.Println("run failed:", err)
	}
}
