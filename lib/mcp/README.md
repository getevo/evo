# MCP

Wire types and JSON Schema generation for the Model Context Protocol.

This package is the protocol layer only: JSON-RPC envelopes, MCP result shapes, protocol revision constants, and a JSON Schema generator that reads Go struct tags. It has no dependency on the root `evo` package, which is what lets the root package import it.

**You normally do not import this package directly.** Tool registration, HTTP transport and authentication live in the root package — see **[docs/mcp.md](../../docs/mcp.md)** for how to expose tools from an application.

## What is here

| Symbol | Purpose |
|---|---|
| `Request`, `Response`, `Error` | JSON-RPC 2.0 envelopes. |
| `Result`, `Failure`, `UnsupportedVersion` | Response constructors. |
| `Content`, `Text` | Tool result content blocks. |
| `CallToolResult`, `ListToolsResult`, `ToolDefinition` | MCP result shapes. |
| `InitializeResult`, `DiscoverResult` | Handshake results for the legacy and modern eras. |
| `Schema`, `GenerateSchema`, `EmptyObjectSchema` | JSON Schema (draft 2020-12) generation from a Go struct. |
| `SupportedVersions`, `IsSupportedVersion`, `IsModern` | Protocol revision helpers. |
| `Code*` constants | JSON-RPC and MCP error codes. |

## Direct use

The one type you may reach for from a tool handler is `CallToolResult`, when you want full control over the result instead of the default shaping:

```go
import "github.com/getevo/evo/v2/lib/mcp"

return &mcp.CallToolResult{
    Content:           []mcp.Content{mcp.Text("2 invoices are overdue")},
    StructuredContent: map[string]int{"overdue": 2},
}
```

`GenerateSchema` is also usable on its own if you need a JSON Schema for a struct outside MCP:

```go
schema := mcp.GenerateSchema(SearchInput{})
```

See [docs/mcp.md](../../docs/mcp.md) for the tag reference it uses.
