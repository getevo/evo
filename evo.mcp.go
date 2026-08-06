package evo

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/mcp"
	"github.com/getevo/evo/v2/lib/outcome"
	"github.com/getevo/evo/v2/lib/validation"
	"github.com/getevo/json"
	"github.com/tidwall/gjson"
)

// MCPConfig configures the built-in MCP (Model Context Protocol) endpoint.
// It is read from the MCP section of the configuration file at Setup() time.
//
// The endpoint is disabled by default: it is a remote procedure call surface
// into the application and must be turned on deliberately.
type MCPConfig struct {
	Enabled bool   `description:"Serve the MCP endpoint" default:"false" json:"enabled" yaml:"enabled"`
	Path    string `description:"HTTP path of the MCP endpoint" default:"/mcp" json:"path" yaml:"path"`
	Name    string `description:"Server name reported to MCP clients" default:"evo" json:"name" yaml:"name"`
	Version string `description:"Server version reported to MCP clients" default:"1.0.0" json:"version" yaml:"version"`

	// Instructions is optional natural language guidance handed to the model
	// describing what this server is for.
	Instructions string `description:"Guidance for the model on how to use this server" default:"" json:"instructions" yaml:"instructions"`

	// Token, when set, must be presented by every request as
	// "Authorization: Bearer <token>". An empty value disables token checking,
	// which is only appropriate behind a trusted gateway.
	Token string `description:"Static bearer token required on every request" default:"" json:"token" yaml:"token"`

	// AllowedOrigins is a comma separated allow list for the Origin header,
	// required by the specification to prevent DNS rebinding attacks. Requests
	// without an Origin header (native MCP clients send none) always pass.
	// "*" allows any origin.
	AllowedOrigins string `description:"Comma separated allowed Origin values, * for any" default:"" json:"allowed_origins" yaml:"allowed_origins"`
}

// mcpConfig holds the effective configuration. Defaults are set here because
// the `default` struct tag is documentation only — nothing reads it back.
var mcpConfig = MCPConfig{
	Path:    "/mcp",
	Name:    "evo",
	Version: "1.0.0",
}

// MCPToolHandler executes a tool call. The value it returns is shaped into an
// MCP tool result:
//
//	nil                    -> empty successful result
//	error / []error        -> tool execution error (isError: true)
//	string / []byte        -> a single text content block
//	*mcp.CallToolResult    -> returned verbatim, for full control
//	anything else          -> structuredContent plus a JSON text mirror
type MCPToolHandler func(c *MCPContext) any

// MCPTool describes one tool exposed over the MCP endpoint.
type MCPTool struct {
	// Name uniquely identifies the tool. Letters, digits, underscore, hyphen
	// and dot only.
	Name string

	// Title is an optional human readable name for display.
	Title string

	// Description tells the model what the tool does. Write it for the model,
	// not for a developer — it is the only thing the model sees when deciding
	// whether to call.
	Description string

	// Input is a zero value of the struct the arguments decode into, for
	// example GetInvoiceInput{}. Its JSON Schema is derived from the `json`,
	// `description`, `default` and `validation` tags. A nil Input means the
	// tool takes no arguments.
	Input any

	// Output optionally describes the result shape the same way.
	Output any

	// Permission, when set, is checked with Request.User().HasPermission
	// before the tool is listed or called. A caller lacking it never sees the
	// tool in tools/list and gets "unknown tool" if it calls anyway.
	Permission string

	// Behavioural hints passed through to the client. They are advisory: a
	// client may use ReadOnly to skip a confirmation prompt, or Destructive to
	// insist on one.
	ReadOnly    bool
	Destructive bool
	Idempotent  bool
	OpenWorld   bool

	// Handler executes the call.
	Handler MCPToolHandler

	inputSchema  *mcp.Schema
	outputSchema *mcp.Schema
}

var (
	mcpTools     = map[string]*MCPTool{}
	mcpToolOrder []string
	mcpMutex     sync.RWMutex
)

// RegisterMCPTool registers one or more tools on the MCP endpoint. It is
// additive and safe to call from any number of sub-applications, normally from
// their Router() method.
//
// Example:
//
//	evo.RegisterMCPTool(evo.MCPTool{
//	    Name:        "get_invoice",
//	    Description: "Fetch an invoice by its identifier",
//	    Input:       GetInvoiceInput{},
//	    ReadOnly:    true,
//	    Handler: func(c *evo.MCPContext) any {
//	        var in GetInvoiceInput
//	        if err := c.Bind(&in); err != nil {
//	            return err
//	        }
//	        return findInvoice(in.ID)
//	    },
//	})
func RegisterMCPTool(tools ...MCPTool) {
	mcpMutex.Lock()
	defer mcpMutex.Unlock()

	for i := range tools {
		tool := tools[i]
		if tool.Name == "" {
			log.Fatalf("mcp: tool registered without a name")
		}
		if tool.Handler == nil {
			log.Fatalf("mcp: tool %s registered without a handler", tool.Name)
		}
		if invalid := strings.TrimLeft(tool.Name, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-."); invalid != "" {
			log.Fatalf("mcp: tool name %q contains unsupported characters", tool.Name)
		}

		tool.inputSchema = mcp.GenerateSchema(tool.Input)
		if tool.Output != nil {
			tool.outputSchema = mcp.GenerateSchema(tool.Output)
		}

		if _, exists := mcpTools[tool.Name]; exists {
			log.Warningf("mcp: tool %s registered twice, replacing the earlier definition", tool.Name)
		} else {
			mcpToolOrder = append(mcpToolOrder, tool.Name)
		}
		mcpTools[tool.Name] = &tool
	}
}

// MCPTools returns the registered tools in registration order.
func MCPTools() []*MCPTool {
	mcpMutex.RLock()
	defer mcpMutex.RUnlock()

	out := make([]*MCPTool, 0, len(mcpToolOrder))
	for _, name := range mcpToolOrder {
		out = append(out, mcpTools[name])
	}
	return out
}

// definition renders the tool as a tools/list entry.
func (t *MCPTool) definition() *mcp.ToolDefinition {
	def := &mcp.ToolDefinition{
		Name:         t.Name,
		Title:        t.Title,
		Description:  t.Description,
		InputSchema:  t.inputSchema,
		OutputSchema: t.outputSchema,
	}
	if t.ReadOnly || t.Destructive || t.Idempotent || t.OpenWorld {
		def.Annotations = &mcp.Annotations{
			ReadOnlyHint:    t.ReadOnly,
			DestructiveHint: t.Destructive,
			IdempotentHint:  t.Idempotent,
			OpenWorldHint:   t.OpenWorld,
		}
	}
	return def
}

// MCPContext is handed to a tool handler. It embeds *Request, so everything
// available to an HTTP handler — headers, client IP, User(), the fiber context
// — is available to a tool as well.
type MCPContext struct {
	*Request

	// Tool is the definition being invoked.
	Tool *MCPTool

	// Arguments is the raw `arguments` object of the call. Prefer Bind.
	Arguments json.RawMessage

	// Client identifies the calling MCP client, when it declared itself.
	Client mcp.Implementation

	// Version is the protocol revision this request declared.
	Version string

	// RequestID is the JSON-RPC id, useful for log correlation.
	RequestID any

	// Ctx is cancelled when the client disconnects. Pass it to long running
	// work, for example db.GetInstance().WithContext(c.Ctx).
	Ctx context.Context
}

// Bind decodes the call arguments into dst and validates it using the
// `validation` struct tags. The returned error is safe to return straight from
// a handler: it becomes a tool execution error the model can correct.
func (c *MCPContext) Bind(dst any) error {
	if len(c.Arguments) > 0 {
		if err := json.Unmarshal(c.Arguments, dst); err != nil {
			return fmt.Errorf("invalid arguments: %w", err)
		}
	}
	if errs := validation.Struct(dst); len(errs) > 0 {
		messages := make([]string, 0, len(errs))
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("invalid arguments: %s", strings.Join(messages, "; "))
	}
	return nil
}

// Arg reads a single argument by dotted path, for example c.Arg("filter.from").
func (c *MCPContext) Arg(path string) generic.Value {
	if len(c.Arguments) == 0 {
		return generic.Parse(nil)
	}
	result := gjson.GetBytes(c.Arguments, path)
	if !result.Exists() {
		return generic.Parse(nil)
	}
	return generic.Parse(result.Value())
}

// Progress reports partial progress of a long running tool call.
//
// This build answers every call with a single JSON response, so there is no
// stream to deliver progress on and the call is recorded at debug level only.
// The method exists now so that enabling streaming later does not change any
// handler signature.
func (c *MCPContext) Progress(percent float64, message string) {
	log.Debugf("mcp: %s progress %.0f%% %s", c.Tool.Name, percent, message)
}

// registerMCPEndpoints mounts the MCP endpoint. It is called from Run() once
// every sub-application has registered its tools, and before the catch-all 404
// middleware is installed.
func registerMCPEndpoints() {
	if !mcpConfig.Enabled {
		return
	}
	if mcpConfig.Path == "" {
		mcpConfig.Path = "/mcp"
	}
	All(mcpConfig.Path, mcpHandler)
	log.Infof("mcp: serving %d tools at %s", len(mcpToolOrder), mcpConfig.Path)
}

// mcpHandler is the single entry point of the MCP endpoint. It writes raw
// JSON-RPC, bypassing the usual success/data envelope, by returning an
// *outcome.Response.
func mcpHandler(r *Request) any {
	// Only POST carries MCP traffic. The GET stream and DELETE session
	// teardown of earlier revisions are not implemented.
	if !strings.EqualFold(r.Method(), "POST") {
		return outcome.Json(mcp.Failure(nil, mcp.CodeInvalidRequest,
			"the MCP endpoint accepts POST only")).Status(StatusMethodNotAllowed)
	}

	// DNS rebinding protection, required for HTTP transports.
	if origin := r.Header("Origin"); origin != "" && !mcpOriginAllowed(origin) {
		return outcome.Json(mcp.Failure(nil, mcp.CodeInvalidRequest,
			"origin not allowed")).Status(StatusForbidden)
	}

	if mcpConfig.Token != "" && !mcpTokenAllowed(r.Header("Authorization")) {
		return outcome.Json(mcp.Failure(nil, mcp.CodeInvalidRequest,
			"unauthorized")).Status(StatusUnauthorized).Header("WWW-Authenticate", "Bearer")
	}

	var req mcp.Request
	if err := json.Unmarshal([]byte(r.Body()), &req); err != nil {
		return outcome.Json(mcp.Failure(nil, mcp.CodeParseError,
			"malformed JSON body")).Status(StatusBadRequest)
	}
	if req.Method == "" {
		return outcome.Json(mcp.Failure(req.ID, mcp.CodeInvalidRequest,
			"missing method")).Status(StatusBadRequest)
	}

	// A notification carries no id and must not be answered with a body.
	if req.IsNotification() {
		return outcome.Text("").Status(StatusAccepted)
	}

	version, failure := mcpResolveVersion(r, &req)
	if failure != nil {
		return outcome.Json(failure).Status(StatusBadRequest)
	}
	if failure := mcpValidateHeaders(r, &req); failure != nil {
		return outcome.Json(failure).Status(StatusBadRequest)
	}

	response, status := mcpDispatch(r, &req, version)
	return outcome.Json(response).Status(status)
}

// mcpResolveVersion determines which protocol revision this request speaks.
//
// An initialize request negotiates it in the body; every other request declares
// it in the MCP-Protocol-Version header or in params._meta. A request that
// declares nothing is treated as the oldest revision, which the specification
// permits for pre-2025-06-18 clients.
func mcpResolveVersion(r *Request, req *mcp.Request) (string, *mcp.Response) {
	if req.Method == mcp.MethodInitialize {
		var params mcp.InitializeParams
		if len(req.Params) > 0 {
			_ = json.Unmarshal(req.Params, &params)
		}
		if mcp.IsSupportedVersion(params.ProtocolVersion) {
			return params.ProtocolVersion, nil
		}
		// Legacy clients cannot fall forward, so answer with the newest
		// handshake revision we speak rather than an error.
		return mcp.LatestLegacyVersion, nil
	}

	version := r.Header(mcp.HeaderProtocolVersion)
	if version == "" {
		var envelope struct {
			Meta *mcp.Meta `json:"_meta"`
		}
		if len(req.Params) > 0 {
			_ = json.Unmarshal(req.Params, &envelope)
		}
		if envelope.Meta != nil {
			version = envelope.Meta.ProtocolVersion
		}
	}
	if version == "" {
		return mcp.FallbackVersion, nil
	}
	if !mcp.IsSupportedVersion(version) {
		return "", mcp.UnsupportedVersion(req.ID, version)
	}
	return version, nil
}

// mcpValidateHeaders enforces that the headers the modern transport mirrors
// from the body actually agree with the body, which stops a proxy and this
// server from acting on different values.
//
// Presence is not required. The specification asks servers to reject a missing
// Mcp-Method or Mcp-Name, but doing so would lock out otherwise correct clients
// for no security gain, so only a genuine mismatch is rejected.
func mcpValidateHeaders(r *Request, req *mcp.Request) *mcp.Response {
	if got := r.Header(mcp.HeaderMethod); got != "" && got != req.Method {
		return mcp.Failure(req.ID, mcp.CodeHeaderMismatch, fmt.Sprintf(
			"Header mismatch: Mcp-Method header value %q does not match body value %q", got, req.Method))
	}

	got := r.Header(mcp.HeaderName)
	if got == "" || req.Method != mcp.MethodToolsCall {
		return nil
	}
	var params mcp.CallToolParams
	if len(req.Params) > 0 {
		_ = json.Unmarshal(req.Params, &params)
	}
	if decoded := mcpDecodeHeader(got); decoded != params.Name {
		return mcp.Failure(req.ID, mcp.CodeHeaderMismatch, fmt.Sprintf(
			"Header mismatch: Mcp-Name header value %q does not match body value %q", decoded, params.Name))
	}
	return nil
}

// mcpDecodeHeader unwraps the =?base64?...?= sentinel the transport uses for
// header values that cannot be written as plain ASCII.
func mcpDecodeHeader(value string) string {
	const prefix, suffix = "=?base64?", "?="
	if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, suffix) {
		return value
	}
	raw, err := base64.StdEncoding.DecodeString(value[len(prefix) : len(value)-len(suffix)])
	if err != nil {
		return value
	}
	return string(raw)
}

// mcpDispatch routes a JSON-RPC request to its handler and reports the HTTP
// status the response should carry.
func mcpDispatch(r *Request, req *mcp.Request, version string) (*mcp.Response, int) {
	modern := mcp.IsModern(version)
	resultType := ""
	if modern {
		resultType = mcp.ResultTypeComplete
	}

	switch req.Method {
	case mcp.MethodInitialize:
		return mcp.Result(req.ID, mcp.InitializeResult{
			ProtocolVersion: version,
			Capabilities:    mcp.Capabilities{Tools: &mcp.ToolsCapability{}},
			ServerInfo:      mcp.Implementation{Name: mcpConfig.Name, Version: mcpConfig.Version},
			Instructions:    mcpConfig.Instructions,
		}), StatusOK

	case mcp.MethodDiscover:
		return mcp.Result(req.ID, mcp.DiscoverResult{
			ResultType:        mcp.ResultTypeComplete,
			SupportedVersions: mcp.SupportedVersions,
			Capabilities:      mcp.Capabilities{Tools: &mcp.ToolsCapability{}},
			Instructions:      mcpConfig.Instructions,
			Meta: &mcp.Meta{
				ServerInfo: &mcp.Implementation{Name: mcpConfig.Name, Version: mcpConfig.Version},
			},
		}), StatusOK

	case mcp.MethodPing:
		return mcp.Result(req.ID, mcp.EmptyResult{ResultType: resultType}), StatusOK

	case mcp.MethodToolsList:
		return mcp.Result(req.ID, mcp.ListToolsResult{
			ResultType: resultType,
			Tools:      mcpVisibleTools(r),
		}), StatusOK

	case mcp.MethodToolsCall:
		return mcpCallTool(r, req, version, resultType)
	}

	// The modern transport asks for 404 on an unknown method so that it is
	// distinguishable from a legacy endpoint that is simply absent.
	status := StatusOK
	if modern {
		status = StatusNotFound
	}
	return mcp.Failure(req.ID, mcp.CodeMethodNotFound,
		fmt.Sprintf("Method not found: %s", req.Method)), status
}

// mcpVisibleTools lists the tools this caller is permitted to see.
func mcpVisibleTools(r *Request) []*mcp.ToolDefinition {
	mcpMutex.RLock()
	defer mcpMutex.RUnlock()

	out := make([]*mcp.ToolDefinition, 0, len(mcpToolOrder))
	for _, name := range mcpToolOrder {
		tool := mcpTools[name]
		if !mcpPermitted(r, tool) {
			continue
		}
		out = append(out, tool.definition())
	}
	return out
}

// mcpPermitted reports whether the caller may see and call the tool.
func mcpPermitted(r *Request, tool *MCPTool) bool {
	if tool.Permission == "" {
		return true
	}
	return r.User().HasPermission(tool.Permission)
}

// mcpCallTool executes a tools/call request.
func mcpCallTool(r *Request, req *mcp.Request, version, resultType string) (*mcp.Response, int) {
	var params mcp.CallToolParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return mcp.Failure(req.ID, mcp.CodeInvalidParams, "malformed params"), StatusOK
		}
	}
	if params.Name == "" {
		return mcp.Failure(req.ID, mcp.CodeInvalidParams, "missing tool name"), StatusOK
	}

	mcpMutex.RLock()
	tool := mcpTools[params.Name]
	mcpMutex.RUnlock()

	// A tool the caller may not use is reported as absent rather than
	// forbidden, so that the endpoint does not disclose what it hides.
	if tool == nil || !mcpPermitted(r, tool) {
		return mcp.Failure(req.ID, mcp.CodeInvalidParams,
			fmt.Sprintf("Unknown tool: %s", params.Name)), StatusOK
	}

	c := &MCPContext{
		Request:   r,
		Tool:      tool,
		Arguments: params.Arguments,
		Version:   version,
		RequestID: req.ID,
		Ctx:       context.Background(),
	}
	if r.Context != nil {
		c.Ctx = r.Context.Context()
	}
	var envelope struct {
		Meta *mcp.Meta `json:"_meta"`
	}
	if len(req.Params) > 0 {
		_ = json.Unmarshal(req.Params, &envelope)
	}
	if envelope.Meta != nil && envelope.Meta.ClientInfo != nil {
		c.Client = *envelope.Meta.ClientInfo
	}

	result := mcpInvoke(c)
	result.ResultType = resultType
	return mcp.Result(req.ID, result), StatusOK
}

// mcpInvoke runs the handler and shapes whatever it returns into a tool result.
// A panicking tool is contained: it becomes a tool execution error rather than
// taking the server down.
func mcpInvoke(c *MCPContext) (result *mcp.CallToolResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Errorf("mcp: tool %s panicked: %v", c.Tool.Name, recovered)
			result = mcpErrorResult(fmt.Sprintf("tool %s failed unexpectedly", c.Tool.Name))
		}
	}()
	return mcpShapeResult(c.Tool.Handler(c))
}

// mcpShapeResult converts a handler return value into a tool result.
func mcpShapeResult(value any) *mcp.CallToolResult {
	switch v := value.(type) {
	case nil:
		return &mcp.CallToolResult{Content: []mcp.Content{}}
	case *mcp.CallToolResult:
		if v == nil {
			return &mcp.CallToolResult{Content: []mcp.Content{}}
		}
		if v.Content == nil {
			v.Content = []mcp.Content{}
		}
		return v
	case error:
		return mcpErrorResult(v.Error())
	case []error:
		messages := make([]string, 0, len(v))
		for _, err := range v {
			messages = append(messages, err.Error())
		}
		return mcpErrorResult(strings.Join(messages, "; "))
	case string:
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.Text(v)}}
	case []byte:
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.Text(string(v))}}
	case []mcp.Content:
		return &mcp.CallToolResult{Content: v}
	case mcp.Content:
		return &mcp.CallToolResult{Content: []mcp.Content{v}}
	}

	// Everything else is structured data. The JSON mirror in content keeps
	// clients that ignore structuredContent working.
	raw, err := json.Marshal(value)
	if err != nil {
		return mcpErrorResult(fmt.Sprintf("tool result could not be encoded: %v", err))
	}
	return &mcp.CallToolResult{
		Content:           []mcp.Content{mcp.Text(string(raw))},
		StructuredContent: value,
	}
}

func mcpErrorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{mcp.Text(message)}, IsError: true}
}

// mcpOriginAllowed checks a present Origin header against the allow list.
func mcpOriginAllowed(origin string) bool {
	list := strings.TrimSpace(mcpConfig.AllowedOrigins)
	if list == "" {
		// Nothing is allow listed, so only clients that send no Origin at all
		// — which is every native MCP client — may connect.
		return false
	}
	for _, allowed := range strings.Split(list, ",") {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" || strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
}

// mcpTokenAllowed compares the Authorization header against the configured
// bearer token in constant time.
func mcpTokenAllowed(header string) bool {
	const prefix = "bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return false
	}
	presented := strings.TrimSpace(header[len(prefix):])
	return subtle.ConstantTimeCompare([]byte(presented), []byte(mcpConfig.Token)) == 1
}
