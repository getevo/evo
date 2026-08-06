// Example application exposing tools to AI agents over MCP.
//
// Run it with an MCP section in config.yml:
//
//	MCP:
//	  enabled: true
//	  path: /mcp
//	  name: invoice-service
//	  version: 1.0.0
//	  instructions: Look up and issue customer invoices.
//	  token: dev-token
//
// Then point an MCP client at http://127.0.0.1:8080/mcp with the bearer token,
// or probe it by hand:
//
//	curl -s http://127.0.0.1:8080/mcp \
//	  -H 'Authorization: Bearer dev-token' \
//	  -H 'Content-Type: application/json' \
//	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/mcp"
)

// --- domain ----------------------------------------------------------------

type Invoice struct {
	ID       string    `json:"id"`
	Customer string    `json:"customer"`
	Total    float64   `json:"total"`
	Status   string    `json:"status"`
	IssuedAt time.Time `json:"issued_at"`
}

// invoices stands in for a database in this example.
var invoices = map[string]*Invoice{
	"INV-1001": {ID: "INV-1001", Customer: "Acme", Total: 240.50, Status: "paid", IssuedAt: time.Now().Add(-48 * time.Hour)},
	"INV-1002": {ID: "INV-1002", Customer: "Globex", Total: 99.00, Status: "draft", IssuedAt: time.Now().Add(-2 * time.Hour)},
}

// --- tool inputs -----------------------------------------------------------

// The JSON Schema the model sees is derived from these tags: `json` gives the
// argument name, `description` explains it, and `validation` becomes both the
// advertised constraint and the check Bind() enforces at call time.
type GetInvoiceInput struct {
	ID string `json:"id" validation:"required,regex(^INV-[0-9]{4}$)" description:"Invoice identifier, for example INV-1001"`
}

// Status and Limit are pointers because they are optional *and* constrained.
// A plain `int` would arrive as 0 when the model omits it, and 0 satisfies
// neither in(draft,sent,paid) nor >=1, so the call would always fail
// validation. A nil pointer is skipped instead.
type SearchInvoicesInput struct {
	Customer string  `json:"customer" description:"Filter by customer name, case-insensitive"`
	Status   *string `json:"status" validation:"in(draft,sent,paid)" description:"Filter by invoice status"`
	Limit    *int    `json:"limit" validation:">=1,<=50" default:"10" description:"Maximum number of invoices to return"`
}

type IssueInvoiceInput struct {
	Customer string  `json:"customer" validation:"required,len<=80" description:"Customer name to bill"`
	Total    float64 `json:"total" validation:"required,>0" description:"Invoice total in euros"`
}

// --- application -----------------------------------------------------------

type App struct{}

func (App) Name() string { return "invoices" }

func (App) Register() error { return nil }

func (App) WhenReady() error { return nil }

// Router registers both HTTP routes and MCP tools. Tools are additive, so any
// number of sub-applications can contribute to the same endpoint.
func (App) Router() error {
	evo.RegisterMCPTool(
		// A read-only lookup. ReadOnly lets a client skip its confirmation
		// prompt, because the call cannot change anything.
		evo.MCPTool{
			Name:        "get_invoice",
			Title:       "Get Invoice",
			Description: "Fetch a single invoice by its identifier. Returns the customer, total and status.",
			Input:       GetInvoiceInput{},
			Output:      Invoice{},
			ReadOnly:    true,
			Idempotent:  true,
			Handler:     getInvoice,
		},

		// A search. Returning a slice becomes structuredContent plus a JSON
		// text mirror, so clients that only read text still work.
		evo.MCPTool{
			Name:        "search_invoices",
			Title:       "Search Invoices",
			Description: "Find invoices by customer or status. Use it when the user does not know the invoice number.",
			Input:       SearchInvoicesInput{},
			ReadOnly:    true,
			Handler:     searchInvoices,
		},

		// A write. Permission gates it: a caller without invoice.write never
		// sees this tool in tools/list and cannot call it.
		evo.MCPTool{
			Name:        "issue_invoice",
			Title:       "Issue Invoice",
			Description: "Create a new draft invoice for a customer.",
			Input:       IssueInvoiceInput{},
			Output:      Invoice{},
			Permission:  "invoice.write",
			Handler:     issueInvoice,
		},

		// A tool with no arguments, and full control over its result.
		evo.MCPTool{
			Name:        "invoice_stats",
			Title:       "Invoice Statistics",
			Description: "Report how many invoices exist in each status.",
			ReadOnly:    true,
			Handler:     invoiceStats,
		},
	)
	return nil
}

// --- handlers --------------------------------------------------------------

// getInvoice shows the common shape: bind, act, return the value.
func getInvoice(c *evo.MCPContext) any {
	var in GetInvoiceInput
	if err := c.Bind(&in); err != nil {
		// Returning an error becomes a tool execution error (isError: true),
		// which the model reads and can correct on its next attempt.
		return err
	}

	invoice, found := invoices[in.ID]
	if !found {
		return fmt.Errorf("no invoice %s exists — use search_invoices to find the right identifier", in.ID)
	}
	// A struct becomes structuredContent plus a JSON text mirror.
	return invoice
}

func searchInvoices(c *evo.MCPContext) any {
	var in SearchInvoicesInput
	if err := c.Bind(&in); err != nil {
		return err
	}

	// The `default` tag documents the schema; applying it is up to the handler.
	limit := 10
	if in.Limit != nil {
		limit = *in.Limit
	}

	var matches []*Invoice
	for _, invoice := range invoices {
		if in.Customer != "" && !strings.Contains(strings.ToLower(invoice.Customer), strings.ToLower(in.Customer)) {
			continue
		}
		if in.Status != nil && invoice.Status != *in.Status {
			continue
		}
		matches = append(matches, invoice)
		if len(matches) >= limit {
			break
		}
	}

	if len(matches) == 0 {
		// A plain string becomes a single text block. Telling the model that
		// nothing matched beats returning an empty array it has to interpret.
		return "No invoices matched that filter."
	}
	return map[string]any{"invoices": matches, "count": len(matches)}
}

func issueInvoice(c *evo.MCPContext) any {
	var in IssueInvoiceInput
	if err := c.Bind(&in); err != nil {
		return err
	}

	// MCPContext embeds *evo.Request, so the caller's identity, headers and IP
	// are all available — the same as in any HTTP handler.
	log.Infof("mcp: %s issuing an invoice for %s", c.User().GetEmail(), in.Customer)

	invoice := &Invoice{
		ID:       fmt.Sprintf("INV-%d", 1003+len(invoices)),
		Customer: in.Customer,
		Total:    in.Total,
		Status:   "draft",
		IssuedAt: time.Now(),
	}
	invoices[invoice.ID] = invoice
	return invoice
}

// invoiceStats returns an *mcp.CallToolResult directly, which is the escape
// hatch when the default shaping is not what you want.
func invoiceStats(c *evo.MCPContext) any {
	counts := map[string]int{}
	for _, invoice := range invoices {
		counts[invoice.Status]++
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("%d invoices total", len(invoices)))
	for status, count := range counts {
		summary.WriteString(fmt.Sprintf("\n  %s: %d", status, count))
	}

	return &mcp.CallToolResult{
		Content:           []mcp.Content{mcp.Text(summary.String())},
		StructuredContent: counts,
	}
}

// --- entry point -----------------------------------------------------------

func main() {
	if err := evo.Setup(); err != nil {
		log.Fatal(err)
	}
	evo.Register(App{})
	if err := evo.Run(); err != nil {
		log.Fatal(err)
	}
}
