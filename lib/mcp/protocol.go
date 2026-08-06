// Package mcp implements the wire protocol of the Model Context Protocol (MCP)
// over Streamable HTTP, independent of any transport or web framework.
//
// The package is deliberately free of dependencies on the root evo package so
// that the root package can import it. Transport wiring, tool registration and
// authentication live in the root package (see evo.mcp.go).
//
// Two protocol eras are supported:
//
//   - legacy  (2025-03-26 .. 2025-11-25) — the client opens with an `initialize`
//     handshake and results carry no `resultType`.
//   - modern  (2026-07-28 and later)     — every request declares its own
//     protocol version and results carry `"resultType": "complete"`.
//
// Both eras are served statelessly on the same endpoint. Session identifiers
// are never minted, no SSE stream is opened and no GET stream is offered, which
// the specification permits for a server that has nothing to stream.
package mcp

import (
	"github.com/getevo/json"
)

// Protocol revisions understood by this implementation, newest first.
const (
	// Version20260728 is the first "modern" revision: stateless, no
	// initialize handshake, per-request version negotiation.
	Version20260728 = "2026-07-28"
	Version20251125 = "2025-11-25"
	Version20250618 = "2025-06-18"
	Version20250326 = "2025-03-26"
)

// LatestVersion is the newest revision this implementation speaks.
const LatestVersion = Version20260728

// LatestLegacyVersion is the newest handshake-based revision this
// implementation speaks. It is used as the fallback when a legacy client asks
// to initialize with a revision we do not recognise.
const LatestLegacyVersion = Version20251125

// FallbackVersion is assumed when a request carries no MCP-Protocol-Version
// header, as permitted by the specification for pre-2025-06-18 clients.
const FallbackVersion = Version20250326

// SupportedVersions lists every revision this implementation accepts, newest
// first. The order is significant: it is reported verbatim to clients.
var SupportedVersions = []string{
	Version20260728,
	Version20251125,
	Version20250618,
	Version20250326,
}

// IsSupportedVersion reports whether version is one this implementation speaks.
func IsSupportedVersion(version string) bool {
	for _, v := range SupportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

// IsModern reports whether version belongs to the modern (stateless,
// per-request metadata) era rather than the legacy handshake era.
func IsModern(version string) bool {
	return version >= Version20260728
}

// Keys used inside the `_meta` object of modern requests and results.
const (
	MetaProtocolVersion  = "io.modelcontextprotocol/protocolVersion"
	MetaClientInfo       = "io.modelcontextprotocol/clientInfo"
	MetaClientCapability = "io.modelcontextprotocol/clientCapabilities"
	MetaServerInfo       = "io.modelcontextprotocol/serverInfo"
)

// RPC method names handled by this implementation.
const (
	MethodInitialize  = "initialize"
	MethodInitialized = "notifications/initialized"
	MethodDiscover    = "server/discover"
	MethodToolsList   = "tools/list"
	MethodToolsCall   = "tools/call"
	MethodPing        = "ping"
)

// HTTP headers defined by the Streamable HTTP transport.
const (
	HeaderProtocolVersion = "MCP-Protocol-Version"
	HeaderMethod          = "Mcp-Method"
	HeaderName            = "Mcp-Name"
	HeaderSessionID       = "Mcp-Session-Id"
)

// JSON-RPC and MCP error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603

	// CodeHeaderMismatch is returned when a mirrored HTTP header disagrees
	// with the corresponding value in the request body.
	CodeHeaderMismatch = -32020
	// CodeUnsupportedProtocolVersion is returned when the client asks for a
	// revision this server does not implement.
	CodeUnsupportedProtocolVersion = -32022
)

// ResultTypeComplete is the discriminator modern results must carry. Legacy
// results omit it entirely.
const ResultTypeComplete = "complete"

// Request is an incoming JSON-RPC 2.0 request or notification.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// IsNotification reports whether the message is a notification, meaning it
// carries no id and therefore must not be answered with a JSON-RPC response.
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// Response is an outgoing JSON-RPC 2.0 response. Exactly one of Result and
// Error is populated.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error is a JSON-RPC 2.0 error object.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Result builds a successful JSON-RPC response.
func Result(id any, result any) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Result: result}
}

// Failure builds a JSON-RPC error response. At most one data value is used.
func Failure(id any, code int, message string, data ...any) *Response {
	e := &Error{Code: code, Message: message}
	if len(data) > 0 {
		e.Data = data[0]
	}
	return &Response{JSONRPC: "2.0", ID: id, Error: e}
}

// UnsupportedVersion builds the error a server must return when it does not
// implement the revision the client asked for.
func UnsupportedVersion(id any, requested string) *Response {
	return Failure(id, CodeUnsupportedProtocolVersion, "Unsupported protocol version", map[string]any{
		"supported": SupportedVersions,
		"requested": requested,
	})
}

// Meta is the `_meta` object carried by modern requests and results.
type Meta struct {
	ProtocolVersion    string          `json:"io.modelcontextprotocol/protocolVersion,omitempty"`
	ClientInfo         *Implementation `json:"io.modelcontextprotocol/clientInfo,omitempty"`
	ClientCapabilities map[string]any  `json:"io.modelcontextprotocol/clientCapabilities,omitempty"`
	ServerInfo         *Implementation `json:"io.modelcontextprotocol/serverInfo,omitempty"`
}

// Implementation identifies a client or server by name and version.
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Title   string `json:"title,omitempty"`
}

// Capabilities advertises the server features a client may rely on.
type Capabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability declares tool support. ListChanged stays false because this
// implementation has no notification stream to deliver changes on.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// InitializeParams are the params of a legacy `initialize` request.
type InitializeParams struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    map[string]any  `json:"capabilities,omitempty"`
	ClientInfo      *Implementation `json:"clientInfo,omitempty"`
}

// InitializeResult answers a legacy `initialize` request.
type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    Capabilities   `json:"capabilities"`
	ServerInfo      Implementation `json:"serverInfo"`
	Instructions    string         `json:"instructions,omitempty"`
}

// DiscoverResult answers a modern `server/discover` request.
type DiscoverResult struct {
	ResultType        string       `json:"resultType"`
	SupportedVersions []string     `json:"supportedVersions"`
	Capabilities      Capabilities `json:"capabilities"`
	Instructions      string       `json:"instructions,omitempty"`
	Meta              *Meta        `json:"_meta,omitempty"`
}

// Annotations are optional behavioural hints attached to a tool definition.
// Clients treat them as untrusted display hints, not as security guarantees.
type Annotations struct {
	Title           string `json:"title,omitempty"`
	ReadOnlyHint    bool   `json:"readOnlyHint,omitempty"`
	DestructiveHint bool   `json:"destructiveHint,omitempty"`
	IdempotentHint  bool   `json:"idempotentHint,omitempty"`
	OpenWorldHint   bool   `json:"openWorldHint,omitempty"`
}

// ToolDefinition is a single entry of a `tools/list` result.
type ToolDefinition struct {
	Name         string       `json:"name"`
	Title        string       `json:"title,omitempty"`
	Description  string       `json:"description,omitempty"`
	InputSchema  *Schema      `json:"inputSchema"`
	OutputSchema *Schema      `json:"outputSchema,omitempty"`
	Annotations  *Annotations `json:"annotations,omitempty"`
}

// ListToolsResult answers a `tools/list` request. ResultType is populated only
// for modern clients.
type ListToolsResult struct {
	ResultType string            `json:"resultType,omitempty"`
	Tools      []*ToolDefinition `json:"tools"`
}

// CallToolParams are the params of a `tools/call` request.
type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// Content is one unstructured content block of a tool result. Only the fields
// relevant to the block's Type are populated.
type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// Text builds a text content block.
func Text(text string) Content {
	return Content{Type: "text", Text: text}
}

// MarshalJSON keeps `text` present on a text block even when it is empty, and
// keeps it absent everywhere else. The field is required on a text block, so
// omitempty must not drop it, but an image or audio block has no text field at
// all.
func (c Content) MarshalJSON() ([]byte, error) {
	if c.Type == "text" {
		return json.Marshal(struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: c.Type, Text: c.Text})
	}
	type block Content // shed this method to avoid recursing
	return json.Marshal(block(c))
}

// CallToolResult answers a `tools/call` request. ResultType is populated only
// for modern clients.
//
// IsError marks a tool execution error: the call reached the tool and the tool
// reported a failure the model can act on. Failures of the request itself
// (unknown tool, malformed params) are JSON-RPC errors instead.
type CallToolResult struct {
	ResultType        string    `json:"resultType,omitempty"`
	Content           []Content `json:"content"`
	StructuredContent any       `json:"structuredContent,omitempty"`
	IsError           bool      `json:"isError,omitempty"`
}

// EmptyResult answers `ping`, which carries no payload.
type EmptyResult struct {
	ResultType string `json:"resultType,omitempty"`
}
