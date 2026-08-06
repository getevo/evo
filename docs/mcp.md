# MCP Server

EVO can expose parts of your application to AI agents as **MCP tools**. The Model Context Protocol is how clients such as Claude Code, Claude Desktop and Cursor discover and call server-side functions. Register a Go function, and an agent can call it.

The endpoint is a single HTTP POST route. It is disabled by default: it is a remote procedure call surface into your application, so you turn it on deliberately.

## Import

```go
import (
    "github.com/getevo/evo/v2"
    "github.com/getevo/evo/v2/lib/mcp"
)
```

`evo` gives you the registration API. `lib/mcp` holds the wire types and is only needed when you want full control over a tool result.

## Quick Start

Enable it in `config.yml`:

```yaml
MCP:
  enabled: true
  path: /mcp
  name: invoice-service
  version: 1.0.0
  token: change-me
```

Register a tool from a sub-application's `Router()`:

```go
type GetInvoiceInput struct {
    ID string `json:"id" validation:"required" description:"Invoice identifier, e.g. INV-1001"`
}

func (App) Router() error {
    evo.RegisterMCPTool(evo.MCPTool{
        Name:        "get_invoice",
        Title:       "Get Invoice",
        Description: "Fetch a single invoice by its identifier.",
        Input:       GetInvoiceInput{},
        ReadOnly:    true,
        Handler: func(c *evo.MCPContext) any {
            var in GetInvoiceInput
            if err := c.Bind(&in); err != nil {
                return err
            }
            return findInvoice(in.ID)
        },
    })
    return nil
}
```

That is the whole integration. The JSON Schema the model sees is derived from the struct tags, arguments are validated before your code runs, and the return value is shaped into an MCP result.

Check it by hand:

```bash
curl -s http://127.0.0.1:8080/mcp \
  -H 'Authorization: Bearer change-me' \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

A complete runnable application is in [`examples/mcp/main.go`](../examples/mcp/main.go).

## Configuration

All keys live under the `MCP` section.

| Key | Default | Description |
|---|---|---|
| `enabled` | `false` | Serve the MCP endpoint. |
| `path` | `/mcp` | HTTP path of the endpoint. |
| `name` | `evo` | Server name reported to clients. |
| `version` | `1.0.0` | Server version reported to clients. |
| `instructions` | `""` | Natural language guidance handed to the model about what this server is for. |
| `token` | `""` | Static bearer token required on every request. Empty disables the check. |
| `allowed_origins` | `""` | Comma separated allow list for the `Origin` header. `*` allows any. |

## How a tool receives its parameters

There is no parameter list to declare and no binding to wire up. You describe the arguments once as a Go struct, and that one struct does three jobs: it tells the model what to send, it validates what arrives, and it is what you read in the handler.

Follow one call end to end.

### 1. Declare the arguments as a struct

```go
type SearchInvoicesInput struct {
    Customer string  `json:"customer" description:"Filter by customer name"`
    Status   *string `json:"status" validation:"in(draft,sent,paid)" description:"Only invoices in this state"`
    Limit    *int    `json:"limit" validation:">=1,<=50" default:"10" description:"How many invoices to return"`
}

evo.RegisterMCPTool(evo.MCPTool{
    Name:        "search_invoices",
    Description: "Find invoices by customer or status.",
    Input:       SearchInvoicesInput{},   // a zero value — only the type is read
    ReadOnly:    true,
    Handler:     searchInvoices,
})
```

`Input` is only inspected for its type, so a zero value is all you pass.

Each tag has a job:

- `json` — the argument name on the wire.
- `description` — what the model reads to decide what to put there.
- `validation` — the constraint, both advertised and enforced.
- `default` — the advertised default. Applying it is up to your handler.

### 2. EVO publishes the schema, and the client shows it to the model

When the client calls `tools/list`, it gets a JSON Schema derived from those tags:

```json
{
  "name": "search_invoices",
  "description": "Find invoices by customer or status.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "customer": { "type": "string", "description": "Filter by customer name" },
      "limit": {
        "type": "integer",
        "description": "How many invoices to return",
        "default": 10,
        "minimum": 1,
        "maximum": 50
      },
      "status": {
        "type": "string",
        "description": "Only invoices in this state",
        "enum": ["draft", "sent", "paid"]
      }
    }
  },
  "annotations": { "readOnlyHint": true }
}
```

This is the only thing the model sees. It has no access to your Go code, your database, or your intent — just these names, types and descriptions.

### 3. The client sends the arguments in `params.arguments`

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "search_invoices",
    "arguments": { "customer": "Acme", "status": "paid", "limit": 5 }
  }
}
```

EVO pulls the `arguments` object out and hands it to your handler as raw JSON on the context.

### 4. Read them with `Bind`

```go
func searchInvoices(c *evo.MCPContext) any {
    var in SearchInvoicesInput
    if err := c.Bind(&in); err != nil {
        return err
    }

    limit := 10                    // the `default` tag documents it; apply it here
    if in.Limit != nil {
        limit = *in.Limit
    }
    if in.Status != nil {
        // ... filter by *in.Status
    }
    return findInvoices(in.Customer, in.Status, limit)
}
```

`Bind` does two things: `json.Unmarshal` into your struct, then `validation.Struct` against the same tags that produced the schema. So the constraint you advertised is the constraint that is actually enforced — they cannot drift apart.

### Reading a single argument without a struct

For a one-argument tool, or to peek at something nested, `Arg` takes a dotted path:

```go
func handler(c *evo.MCPContext) any {
    id    := c.Arg("id").String()
    limit := c.Arg("limit").Int()
    from, err := c.Arg("filter.from").Time()   // nested path; Time returns an error
    if err != nil {
        return fmt.Errorf("filter.from is not a valid date: %w", err)
    }
    return lookup(id, limit, from)
}
```

`String()`, `Int()`, `Float()` and `Bool()` return a single value; `Time()` and `Duration()` also return an error, because parsing can fail.

`Arg` reads the raw JSON directly and **does not validate**. It returns an empty value for a path that is not there, so it never panics on a missing argument. Use `Bind` when you want the constraints enforced; use `Arg` for quick reads.

### Nested and repeated arguments

Nested structs and slices need nothing special — they become nested objects and arrays:

```go
type ReportInput struct {
    Customers []string `json:"customers" validation:"required,min_items(1)" description:"Customer names to include"`
    Period    struct {
        From time.Time `json:"from" validation:"required" description:"Start of the period"`
        To   time.Time `json:"to" validation:"required" description:"End of the period"`
    } `json:"period" description:"Reporting window"`
}
```

```json
"arguments": {
  "customers": ["Acme", "Globex"],
  "period": { "from": "2026-01-01T00:00:00Z", "to": "2026-03-31T00:00:00Z" }
}
```

`time.Time` is advertised as a `date-time` string. Embedded structs are flattened into the parent, matching `encoding/json`.

### When the arguments are wrong

Returning the error from `Bind` sends the reason back to the model as a tool execution error, and the model gets to try again:

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [{
      "type": "text",
      "text": "invalid arguments: status value must be one of: draft,sent,paid; limit is bigger than or equal to 50"
    }],
    "isError": true
  }
}
```

That round trip is why `description` is worth writing on every field. A model that can see *why* the call failed usually fixes it on the next attempt; one that gets "invalid input" guesses again.

### A tool with no parameters

Leave `Input` unset:

```go
evo.RegisterMCPTool(evo.MCPTool{
    Name:        "invoice_stats",
    Description: "Report how many invoices exist in each status.",
    ReadOnly:    true,
    Handler:     func(c *evo.MCPContext) any { return countByStatus() },
})
```

EVO advertises `{"type":"object","additionalProperties":false}`, which is how the specification says "this tool takes nothing".

The full tag reference is in [Input schemas from struct tags](#input-schemas-from-struct-tags) below.

## `evo.RegisterMCPTool(tools ...MCPTool)`

Registers one or more tools. It is additive and safe to call from any number of sub-applications, so each domain package can contribute its own tools to one shared endpoint.

Call it from `Router()`. Registration must finish before the server starts listening; the endpoint is mounted after every sub-application has run.

An invalid registration — no name, no handler, or a name with unsupported characters — is a programming error and aborts startup.

### `MCPTool` fields

| Field | Type | Description |
|---|---|---|
| `Name` | `string` | Unique identifier. Letters, digits, `_`, `-` and `.` only. |
| `Title` | `string` | Optional human readable name for display. |
| `Description` | `string` | What the tool does. **Write this for the model** — it is the only thing the model sees when deciding whether to call. |
| `Input` | `any` | Zero value of the argument struct, e.g. `GetInvoiceInput{}`. `nil` means the tool takes no arguments. |
| `Output` | `any` | Optional zero value describing the result shape. See the note below. |
| `Permission` | `string` | Checked with `Request.User().HasPermission` before the tool is listed or called. |
| `ReadOnly` | `bool` | The call changes nothing. Clients may skip their confirmation prompt. |
| `Destructive` | `bool` | The call may delete or overwrite data. Clients should insist on confirmation. |
| `Idempotent` | `bool` | Repeating the call with the same arguments has no additional effect. |
| `OpenWorld` | `bool` | The tool reaches an external system whose contents are not known up front. |
| `Handler` | `MCPToolHandler` | Executes the call. |

The four hints are advisory. A client uses them to decide how much friction to put in front of a call; they are not enforcement.

### `Output` is a promise you make, not one EVO keeps

Setting `Output` publishes an `outputSchema`, and the specification says a server **must** then return structured results that conform to it. EVO generates and advertises that schema but does **not** check your return value against it. If you declare `Output: Invoice{}` and return a `map[string]any` with different fields, you have published a false schema and a validating client may reject the result.

Either keep the handler's return type matching `Output`, or leave `Output` unset — an absent schema makes no claim at all.

### `evo.MCPTools() []*MCPTool`

Returns the registered tools in registration order. Useful for diagnostics and for your own admin pages.

## Input schemas from struct tags

The `Input` struct is turned into a JSON Schema using tags the framework already uses, so there is nothing new to learn:

| Tag | Purpose |
|---|---|
| `json` | Argument name. `-` omits the field. |
| `description` | Explains the argument to the model. Worth writing for every field. |
| `default` | Advertised default value. |
| `validation` | Constraints — advertised in the schema *and* enforced by `Bind`. |

Validation rules map onto JSON Schema like this:

| Rule | Schema keyword |
|---|---|
| `required` | added to `required` |
| `email`, `url`, `uuid`, `date`, `ipv4`, `ipv6`, `domain` | `format` |
| `regex(...)` | `pattern` |
| `alpha`, `alphanumeric`, `digit`, `slug`, `hex` | `pattern` |
| `in(a,b,c)` | `enum`, typed to the property (see below) |
| `len>N`, `len>=N`, `len<N`, `len<=N`, `len==N` | `minLength` / `maxLength`, or `minItems` / `maxItems` on a slice |
| `>N`, `>=N`, `<N`, `<=N`, `==N` | `minimum` / `maximum` / `exclusiveMinimum` / `exclusiveMaximum` |
| `int`, `+int`, `float`, `+float` | `type`, plus a zero bound for the signed forms |
| `min_items(N)`, `max_items(N)`, `unique_items` | `minItems` / `maxItems` / `uniqueItems` |

Rules with no schema equivalent are simply not advertised. Validation still enforces them at call time — the schema only describes what a client can check before it calls.

`enum` members and `default` values are emitted with the property's own type, not as strings. `Level int` with `in(1,2,3)` produces `"enum": [1, 2, 3]`, and `Limit int` with `default:"10"` produces `"default": 10`. Strings there would be unsatisfiable on a numeric property, so this is deliberate — do not "correct" it back to quoted values.

Go types map as you would expect: strings, numbers, booleans, slices to arrays, nested structs to nested objects, `time.Time` to a `date-time` string. Embedded structs are flattened into the parent, matching `encoding/json`.

A tool with no `Input` advertises `{"type":"object","additionalProperties":false}`, which is the correct way to say "no arguments".

### Optional fields that are also constrained — use a pointer

This one is easy to get wrong. A field the model may omit, which also carries a constraint, must be a **pointer**:

```go
type SearchInput struct {
    Status *string `json:"status" validation:"in(draft,sent,paid)"`
    Limit  *int    `json:"limit" validation:">=1,<=50"`
}
```

With a plain `int`, an omitted `limit` arrives as `0`, and `0` does not satisfy `>=1` — so every call that leaves it out fails validation. A nil pointer is skipped instead, while `required` still fires for genuinely missing values. This is how `lib/validation` behaves everywhere in the framework, not something specific to MCP.

A plain value type is fine when the field is required, or when its zero value is acceptable.

## `MCPContext`

`MCPContext` embeds `*evo.Request`, so everything an HTTP handler can reach — headers, client IP, `User()`, the fiber context — is available inside a tool.

| Field | Type | Description |
|---|---|---|
| `*evo.Request` | embedded | The underlying HTTP request. |
| `Tool` | `*MCPTool` | The definition being invoked. |
| `Arguments` | `json.RawMessage` | The raw `arguments` object. Prefer `Bind`. |
| `Client` | `mcp.Implementation` | Name and version of the calling client, when it declared itself. |
| `Version` | `string` | Protocol revision this request declared. |
| `RequestID` | `any` | The JSON-RPC id, for log correlation. |
| `Ctx` | `context.Context` | Cancelled when the client disconnects. |

### `c.Bind(dst any) error`

Decodes the arguments into `dst` and validates them against the `validation` tags. The returned error is safe to return straight from a handler — it becomes a tool execution error the model can read and correct.

```go
var in GetInvoiceInput
if err := c.Bind(&in); err != nil {
    return err
}
```

### `c.Arg(path string) generic.Value`

Reads a single argument by dotted path, without a struct.

```go
from := c.Arg("filter.from").String()
limit := c.Arg("limit").Int()
```

### `c.Ctx`

Pass it to anything long running so that a disconnecting client actually cancels the work:

```go
db.GetInstance().WithContext(c.Ctx).Find(&rows)
```

### `c.Progress(percent float64, message string)`

Reports partial progress. Every call is currently answered with a single JSON response, so there is no stream to deliver progress on and the call is recorded at debug level. The method exists now so that enabling streaming later will not change any handler signature.

## Handler return values

Whatever the handler returns is shaped into a tool result:

| Return | Result |
|---|---|
| `nil` | Empty successful result. |
| `error` or `[]error` | Tool execution error, `isError: true`. |
| `string` or `[]byte` | One text content block. |
| `mcp.Content` or `[]mcp.Content` | Those content blocks verbatim, for images and audio. |
| `*mcp.CallToolResult` | Returned untouched — full control. |
| anything else | `structuredContent`, plus a JSON text mirror for clients that read only text. |

The JSON mirror matters: not every client reads `structuredContent`, and the specification asks servers to provide both.

### Two kinds of failure

The distinction is worth getting right, because it decides whether the model can recover.

**Tool execution errors** are things the model can fix — a bad date, an unknown identifier, a business rule that says no. Return an `error` and the model sees the message and can try again:

```go
return fmt.Errorf("no invoice %s exists — use search_invoices to find the right identifier", in.ID)
```

Write these messages for the model. "Invalid input" wastes a round trip; naming the problem and the way out does not.

**Protocol errors** are problems with the request itself — unknown tool, malformed params. EVO produces these for you as JSON-RPC errors.

A panicking tool is contained: it becomes a tool execution error and is logged, with the panic detail kept out of the response.

## Authentication and authorization

Two independent layers.

### Layer 1 — reaching the endpoint

Runs before any dispatch, so no tool code executes:

- **Origin validation.** A request whose `Origin` header is not in `allowed_origins` gets `403`. This is required for HTTP transports to prevent DNS rebinding attacks. Requests with no `Origin` at all — which is every native MCP client — always pass.
- **Bearer token.** When `token` is set, every request must carry `Authorization: Bearer <token>` or it gets `401`. The comparison is constant time.

### Layer 2 — who the caller is

`c.User()` goes through your application's `evo.UserInterface`, exactly as in an HTTP handler. Setting `Permission` on a tool gates it:

```go
evo.MCPTool{
    Name:       "issue_invoice",
    Permission: "invoice.write",
    ...
}
```

A caller lacking the permission never sees the tool in `tools/list`, and gets `Unknown tool` if it calls anyway — so the endpoint does not disclose what it hides.

Two things to check before relying on this:

1. **Does your `UserInterface.FromRequest` read `Authorization: Bearer`?** MCP clients send exactly that. If your implementation parses bearer tokens, per-user permissions work with no extra code. If it is cookie or session based, MCP callers will always be anonymous and every permissioned tool will be invisible.
2. **What is `User()` for a machine token?** A local MCP client has no logged-in human. Either map the token to a service user in your `UserInterface`, or leave permissioned tools off the MCP surface.

OAuth 2.1 as a resource server is the specification's answer for public deployments. It is not implemented here; the static token is intended for a trusted network or a gateway that terminates auth.

## Protocol support

| Aspect | Support |
|---|---|
| Transport | Streamable HTTP, POST only |
| Revisions | `2026-07-28`, `2025-11-25`, `2025-06-18`, `2025-03-26` |
| Methods | `initialize`, `notifications/initialized`, `server/discover`, `tools/list`, `tools/call`, `ping` |
| Features | Tools |

Both protocol eras are served on the same endpoint, statelessly:

- **Legacy clients** (`2025-11-25` and earlier) open with an `initialize` handshake. Most clients shipping today do this.
- **Modern clients** (`2026-07-28`) declare their version on every request and may call `server/discover`.

Session identifiers are never minted, no SSE stream is opened, and no GET stream is offered — the specification permits a server with nothing to stream to answer every POST with plain JSON. `GET` and `DELETE` on the endpoint return `405`.

Resources, prompts, sampling, elicitation and progress streaming are not implemented.

### One deliberate deviation

The modern transport mirrors `Mcp-Method` and `Mcp-Name` into HTTP headers and asks servers to reject a request that omits them. EVO validates them **only when present**: a mismatch is rejected with `-32020`, a missing header is not. Requiring them would lock out otherwise correct clients for no security gain, since the body is still the only thing acted upon.

## Notes

- Tools are listed in registration order, which keeps client-side caching and model prompt caches stable.
- Registering the same name twice replaces the earlier definition and logs a warning.
- The endpoint bypasses the usual `{success, data}` envelope — MCP clients need raw JSON-RPC.
- When running locally, bind to `127.0.0.1` rather than `0.0.0.0`.

## See Also

- [Configuration and Settings](configuration.md)
- [Validations](validation.md)
- [Web Server](webserver.md)
- [Health Checks](health-checks.md)
