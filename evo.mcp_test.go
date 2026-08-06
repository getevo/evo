package evo

import (
	"fmt"
	"io"
	nethttp "net/http" // aliased: the package already has an `http` config var
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getevo/evo/v2/lib/mcp"
	"github.com/getevo/json"
	"github.com/gofiber/fiber/v3"
)

// --- harness ---------------------------------------------------------------

// newMCPTestApp builds a bare fiber app with the MCP endpoint mounted and an
// empty tool registry, restoring the previous configuration afterwards.
func newMCPTestApp(t *testing.T) *fiber.App {
	t.Helper()

	a := newTestApp()

	previousConfig := mcpConfig
	mcpMutex.Lock()
	previousTools, previousOrder := mcpTools, mcpToolOrder
	mcpTools, mcpToolOrder = map[string]*MCPTool{}, nil
	mcpMutex.Unlock()

	t.Cleanup(func() {
		mcpConfig = previousConfig
		mcpMutex.Lock()
		mcpTools, mcpToolOrder = previousTools, previousOrder
		mcpMutex.Unlock()
	})

	mcpConfig = MCPConfig{Enabled: true, Path: "/mcp", Name: "test-server", Version: "9.9.9"}
	All("/mcp", mcpHandler)
	return a
}

type rpcResult struct {
	status int
	body   string
	result map[string]any
	err    map[string]any
}

// call posts a JSON-RPC body to /mcp and decodes the envelope.
func call(t *testing.T, a *fiber.App, body string, headers ...string) rpcResult {
	t.Helper()

	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	return do(t, a, req)
}

func do(t *testing.T, a *fiber.App, req *nethttp.Request) rpcResult {
	t.Helper()

	resp, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	out := rpcResult{status: resp.StatusCode, body: string(raw)}
	if len(raw) == 0 {
		return out
	}

	var envelope struct {
		JSONRPC string         `json:"jsonrpc"`
		Result  map[string]any `json:"result"`
		Error   map[string]any `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return out // not JSON, the caller inspects body directly
	}
	if envelope.JSONRPC != "" && envelope.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %q", envelope.JSONRPC)
	}
	out.result, out.err = envelope.Result, envelope.Error
	return out
}

func (r rpcResult) code(t *testing.T) int {
	t.Helper()
	if r.err == nil {
		t.Fatalf("expected a JSON-RPC error, got body %s", r.body)
	}
	code, ok := r.err["code"].(float64)
	if !ok {
		t.Fatalf("error object has no numeric code: %v", r.err)
	}
	return int(code)
}

// echoTool registers a tool that reports the arguments it received.
type echoInput struct {
	Message string `json:"message" validation:"required,len<=20" description:"text to echo"`
	Times   int    `json:"times" validation:">=1,<=3"`
}

func registerEchoTool(t *testing.T) {
	t.Helper()
	RegisterMCPTool(MCPTool{
		Name:        "echo",
		Title:       "Echo",
		Description: "Repeat a message",
		Input:       echoInput{},
		ReadOnly:    true,
		Handler: func(c *MCPContext) any {
			var in echoInput
			if err := c.Bind(&in); err != nil {
				return err
			}
			return map[string]any{
				"echoed": strings.TrimSpace(strings.Repeat(in.Message+" ", in.Times)),
				"times":  in.Times,
			}
		},
	})
}

// --- registration ----------------------------------------------------------

func TestRegisterMCPToolIsAdditiveAndOrdered(t *testing.T) {
	newMCPTestApp(t)

	for _, name := range []string{"first", "second", "third"} {
		RegisterMCPTool(MCPTool{Name: name, Handler: func(c *MCPContext) any { return "ok" }})
	}

	tools := MCPTools()
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}
	for i, want := range []string{"first", "second", "third"} {
		if tools[i].Name != want {
			t.Errorf("position %d: expected %q, got %q", i, want, tools[i].Name)
		}
	}
}

func TestRegisterMCPToolDerivesInputSchema(t *testing.T) {
	newMCPTestApp(t)
	registerEchoTool(t)

	schema := MCPTools()[0].inputSchema
	if schema == nil || schema.Properties["message"] == nil {
		t.Fatalf("expected a derived input schema, got %+v", schema)
	}
	if schema.Properties["message"].Description != "text to echo" {
		t.Error("description tag should reach the schema")
	}
	if len(schema.Required) != 1 || schema.Required[0] != "message" {
		t.Errorf("expected message to be required, got %v", schema.Required)
	}
}

func TestRegisterMCPToolReplacesDuplicateName(t *testing.T) {
	newMCPTestApp(t)

	RegisterMCPTool(MCPTool{Name: "dup", Description: "first", Handler: func(c *MCPContext) any { return "a" }})
	RegisterMCPTool(MCPTool{Name: "dup", Description: "second", Handler: func(c *MCPContext) any { return "b" }})

	tools := MCPTools()
	if len(tools) != 1 {
		t.Fatalf("a duplicate name must not add a second entry, got %d", len(tools))
	}
	if tools[0].Description != "second" {
		t.Errorf("expected the later definition to win, got %q", tools[0].Description)
	}
}

// --- transport -------------------------------------------------------------

func TestMCPRejectsNonPost(t *testing.T) {
	a := newMCPTestApp(t)

	for _, method := range []string{"GET", "DELETE", "PUT"} {
		res := do(t, a, httptest.NewRequest(method, "/mcp", nil))
		if res.status != nethttp.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, res.status)
		}
	}
}

func TestMCPRejectsMalformedBody(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, "{not json")
	if res.status != nethttp.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.status)
	}
	if got := res.code(t); got != mcp.CodeParseError {
		t.Errorf("expected parse error %d, got %d", mcp.CodeParseError, got)
	}
}

func TestMCPRejectsMissingMethod(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1}`)
	if res.status != nethttp.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.status)
	}
	if got := res.code(t); got != mcp.CodeInvalidRequest {
		t.Errorf("expected invalid request %d, got %d", mcp.CodeInvalidRequest, got)
	}
}

func TestMCPNotificationIsAcceptedWithoutBody(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	if res.status != nethttp.StatusAccepted {
		t.Errorf("expected 202, got %d", res.status)
	}
	if strings.TrimSpace(res.body) != "" {
		t.Errorf("a notification must not be answered with a body, got %q", res.body)
	}
}

func TestMCPUnknownMethod(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"resources/list"}`)
	if got := res.code(t); got != mcp.CodeMethodNotFound {
		t.Errorf("expected method not found %d, got %d", mcp.CodeMethodNotFound, got)
	}
	// A legacy client gets the error inside a 200; only modern clients get 404.
	if res.status != nethttp.StatusOK {
		t.Errorf("expected 200 for a legacy client, got %d", res.status)
	}

	res = call(t, a, `{"jsonrpc":"2.0","id":1,"method":"resources/list"}`,
		mcp.HeaderProtocolVersion, mcp.Version20260728)
	if res.status != nethttp.StatusNotFound {
		t.Errorf("expected 404 for a modern client, got %d", res.status)
	}
}

// --- version negotiation ---------------------------------------------------

func TestMCPInitializeEchoesSupportedVersion(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{
		"protocolVersion":"2025-06-18",
		"clientInfo":{"name":"probe","version":"1.2.3"}}}`)

	if res.status != nethttp.StatusOK {
		t.Fatalf("expected 200, got %d body %s", res.status, res.body)
	}
	if got := res.result["protocolVersion"]; got != mcp.Version20250618 {
		t.Errorf("expected the requested version to be echoed, got %v", got)
	}
	if _, present := res.result["resultType"]; present {
		t.Error("a legacy result must not carry resultType")
	}
	info := res.result["serverInfo"].(map[string]any)
	if info["name"] != "test-server" || info["version"] != "9.9.9" {
		t.Errorf("unexpected serverInfo %v", info)
	}
	caps := res.result["capabilities"].(map[string]any)
	if _, ok := caps["tools"]; !ok {
		t.Error("the tools capability must be declared")
	}
}

func TestMCPInitializeFallsBackForUnknownVersion(t *testing.T) {
	a := newMCPTestApp(t)

	// A legacy client cannot fall forward, so it gets our newest legacy
	// revision rather than an error.
	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1999-01-01"}}`)
	if res.status != nethttp.StatusOK {
		t.Fatalf("expected 200, got %d", res.status)
	}
	if got := res.result["protocolVersion"]; got != mcp.LatestLegacyVersion {
		t.Errorf("expected fallback to %s, got %v", mcp.LatestLegacyVersion, got)
	}
}

func TestMCPUnsupportedVersionHeaderIsRejected(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`,
		mcp.HeaderProtocolVersion, "1999-01-01")

	if res.status != nethttp.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.status)
	}
	if got := res.code(t); got != mcp.CodeUnsupportedProtocolVersion {
		t.Errorf("expected %d, got %d", mcp.CodeUnsupportedProtocolVersion, got)
	}
	data := res.err["data"].(map[string]any)
	if data["requested"] != "1999-01-01" {
		t.Errorf("the error must report what was requested, got %v", data)
	}
	if len(data["supported"].([]any)) != len(mcp.SupportedVersions) {
		t.Error("the error must list every supported version")
	}
}

func TestMCPVersionFromRequestMeta(t *testing.T) {
	a := newMCPTestApp(t)

	// A modern client that declares its version only in _meta must still be
	// recognised as modern.
	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{"_meta":{
		"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}`)

	if res.result["resultType"] != mcp.ResultTypeComplete {
		t.Errorf("expected resultType complete, got %v", res.result["resultType"])
	}
}

func TestMCPMissingVersionHeaderIsTreatedAsLegacy(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	if res.status != nethttp.StatusOK {
		t.Fatalf("expected 200, got %d", res.status)
	}
	if _, present := res.result["resultType"]; present {
		t.Error("a legacy result must not carry resultType")
	}
}

func TestMCPDiscover(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":"d1","method":"server/discover","params":{"_meta":{
		"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}`,
		mcp.HeaderProtocolVersion, mcp.Version20260728)

	if res.status != nethttp.StatusOK {
		t.Fatalf("expected 200, got %d body %s", res.status, res.body)
	}
	if res.result["resultType"] != mcp.ResultTypeComplete {
		t.Errorf("expected resultType complete, got %v", res.result["resultType"])
	}
	versions := res.result["supportedVersions"].([]any)
	if len(versions) == 0 || versions[0] != mcp.LatestVersion {
		t.Errorf("expected the newest version first, got %v", versions)
	}
	meta := res.result["_meta"].(map[string]any)
	info := meta[mcp.MetaServerInfo].(map[string]any)
	if info["name"] != "test-server" {
		t.Errorf("expected serverInfo in _meta, got %v", meta)
	}
}

func TestMCPPing(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":7,"method":"ping"}`)
	if res.status != nethttp.StatusOK || res.err != nil {
		t.Fatalf("expected an empty success, got %d %s", res.status, res.body)
	}
}

// --- header mirroring ------------------------------------------------------

func TestMCPHeaderMismatchIsRejected(t *testing.T) {
	a := newMCPTestApp(t)
	registerEchoTool(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`,
		mcp.HeaderMethod, "tools/call")
	if got := res.code(t); got != mcp.CodeHeaderMismatch {
		t.Errorf("expected header mismatch %d, got %d", mcp.CodeHeaderMismatch, got)
	}

	res = call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hi","times":1}}}`,
		mcp.HeaderMethod, "tools/call", mcp.HeaderName, "other")
	if got := res.code(t); got != mcp.CodeHeaderMismatch {
		t.Errorf("expected header mismatch on Mcp-Name, got %d", got)
	}
}

func TestMCPMatchingHeadersPass(t *testing.T) {
	a := newMCPTestApp(t)
	registerEchoTool(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hi","times":1}}}`,
		mcp.HeaderMethod, "tools/call", mcp.HeaderName, "echo")
	if res.err != nil {
		t.Fatalf("matching headers must pass, got %v", res.err)
	}
}

func TestMCPBase64HeaderIsDecodedBeforeComparing(t *testing.T) {
	a := newMCPTestApp(t)
	registerEchoTool(t)

	// "echo" wrapped in the transport's base64 sentinel.
	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hi","times":1}}}`,
		mcp.HeaderName, "=?base64?ZWNobw==?=")
	if res.err != nil {
		t.Fatalf("a base64 encoded header must be decoded before comparing, got %v", res.err)
	}
}

// --- tools/list ------------------------------------------------------------

func TestMCPToolsList(t *testing.T) {
	a := newMCPTestApp(t)
	registerEchoTool(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	tools := res.result["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	tool := tools[0].(map[string]any)
	if tool["name"] != "echo" || tool["title"] != "Echo" || tool["description"] != "Repeat a message" {
		t.Errorf("unexpected tool metadata %v", tool)
	}

	schema := tool["inputSchema"].(map[string]any)
	if schema["type"] != "object" {
		t.Errorf("inputSchema must be an object schema, got %v", schema)
	}
	props := schema["properties"].(map[string]any)
	message := props["message"].(map[string]any)
	if message["type"] != "string" || message["maxLength"].(float64) != 20 {
		t.Errorf("unexpected message schema %v", message)
	}
	times := props["times"].(map[string]any)
	if times["minimum"].(float64) != 1 || times["maximum"].(float64) != 3 {
		t.Errorf("unexpected times bounds %v", times)
	}

	annotations := tool["annotations"].(map[string]any)
	if annotations["readOnlyHint"] != true {
		t.Errorf("expected readOnlyHint, got %v", annotations)
	}
}

func TestMCPToolsListEmptyRegistry(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	tools, ok := res.result["tools"].([]any)
	if !ok {
		t.Fatalf("tools must always be an array, got %T in %s", res.result["tools"], res.body)
	}
	if len(tools) != 0 {
		t.Errorf("expected no tools, got %d", len(tools))
	}
}

func TestMCPToolWithoutInputAdvertisesEmptyObject(t *testing.T) {
	a := newMCPTestApp(t)
	RegisterMCPTool(MCPTool{Name: "now", Handler: func(c *MCPContext) any { return "2026-08-06" }})

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	schema := res.result["tools"].([]any)[0].(map[string]any)["inputSchema"].(map[string]any)
	if schema["type"] != "object" {
		t.Errorf("expected an object schema, got %v", schema)
	}
	if schema["additionalProperties"] != false {
		t.Errorf("a tool with no arguments should accept only an empty object, got %v", schema)
	}
}

// --- tools/call ------------------------------------------------------------

func TestMCPCallToolStructuredResult(t *testing.T) {
	a := newMCPTestApp(t)
	registerEchoTool(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{
		"name":"echo","arguments":{"message":"hi","times":2}}}`)

	if res.err != nil {
		t.Fatalf("unexpected error %v", res.err)
	}
	if res.result["isError"] == true {
		t.Fatalf("unexpected tool error: %s", res.body)
	}

	structured := res.result["structuredContent"].(map[string]any)
	if structured["echoed"] != "hi hi" {
		t.Errorf("unexpected structuredContent %v", structured)
	}

	// Clients that ignore structuredContent must still see the JSON mirror.
	content := res.result["content"].([]any)
	if len(content) != 1 {
		t.Fatalf("expected one content block, got %d", len(content))
	}
	block := content[0].(map[string]any)
	if block["type"] != "text" || !strings.Contains(block["text"].(string), `"echoed"`) {
		t.Errorf("expected a JSON text mirror, got %v", block)
	}
}

func TestMCPCallToolStringResult(t *testing.T) {
	a := newMCPTestApp(t)
	RegisterMCPTool(MCPTool{Name: "greet", Handler: func(c *MCPContext) any { return "hello" }})

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"greet"}}`)
	block := res.result["content"].([]any)[0].(map[string]any)
	if block["type"] != "text" || block["text"] != "hello" {
		t.Errorf("unexpected content %v", block)
	}
	if _, present := res.result["structuredContent"]; present {
		t.Error("a string result should not produce structuredContent")
	}
}

func TestMCPCallToolNilResult(t *testing.T) {
	a := newMCPTestApp(t)
	RegisterMCPTool(MCPTool{Name: "noop", Handler: func(c *MCPContext) any { return nil }})

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"noop"}}`)
	content, ok := res.result["content"].([]any)
	if !ok {
		t.Fatalf("content must always be present, got %s", res.body)
	}
	if len(content) != 0 {
		t.Errorf("expected empty content, got %v", content)
	}
	if res.result["isError"] == true {
		t.Error("nil is a success, not an error")
	}
}

func TestMCPCallToolValidationFailureIsToolError(t *testing.T) {
	a := newMCPTestApp(t)
	registerEchoTool(t)

	// message is required and times must be at least 1.
	res := call(t, a, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{
		"name":"echo","arguments":{"times":9}}}`)

	// A bad argument is something the model can fix, so it must arrive as a
	// tool execution error, not as a JSON-RPC error.
	if res.err != nil {
		t.Fatalf("expected a tool error, got a protocol error %v", res.err)
	}
	if res.result["isError"] != true {
		t.Fatalf("expected isError true, got %s", res.body)
	}
	text := res.result["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(strings.ToLower(text), "invalid arguments") {
		t.Errorf("expected the reason to reach the model, got %q", text)
	}
}

// optionalInput mixes a required field with optional-and-constrained pointer
// fields, which is the shape a tool needs when the model may omit an argument.
type optionalInput struct {
	Name   string  `json:"name" validation:"required"`
	Status *string `json:"status" validation:"in(draft,paid)"`
	Limit  *int    `json:"limit" validation:">=1,<=50"`
}

func TestMCPCallToolOmittedOptionalFieldsPass(t *testing.T) {
	a := newMCPTestApp(t)

	var got optionalInput
	RegisterMCPTool(MCPTool{Name: "search", Input: optionalInput{}, Handler: func(c *MCPContext) any {
		if err := c.Bind(&got); err != nil {
			return err
		}
		return "ok"
	}})

	// A pointer field the caller omitted must not be validated: 0 satisfies
	// neither in(draft,paid) nor >=1, so a value type here would always fail.
	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{
		"name":"search","arguments":{"name":"acme"}}}`)

	if res.result["isError"] == true {
		t.Fatalf("omitting an optional constrained field must not fail: %s", res.body)
	}
	if got.Status != nil || got.Limit != nil {
		t.Errorf("expected the omitted fields to stay nil, got %+v", got)
	}
}

func TestMCPCallToolSuppliedOptionalFieldsAreValidated(t *testing.T) {
	a := newMCPTestApp(t)
	RegisterMCPTool(MCPTool{Name: "search", Input: optionalInput{}, Handler: func(c *MCPContext) any {
		var in optionalInput
		if err := c.Bind(&in); err != nil {
			return err
		}
		return map[string]any{"limit": *in.Limit}
	}})

	// A value the caller did supply is still fully checked.
	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{
		"name":"search","arguments":{"name":"acme","limit":900}}}`)
	if res.result["isError"] != true {
		t.Fatalf("a supplied out-of-range value must fail: %s", res.body)
	}

	res = call(t, a, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{
		"name":"search","arguments":{"name":"acme","limit":10}}}`)
	if res.result["isError"] == true {
		t.Fatalf("a supplied in-range value must pass: %s", res.body)
	}

	// required still fires when the field is genuinely missing.
	res = call(t, a, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{
		"name":"search","arguments":{"limit":10}}}`)
	if res.result["isError"] != true {
		t.Fatalf("required must still be enforced: %s", res.body)
	}
}

func TestMCPCallUnknownTool(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"nope"}}`)
	if got := res.code(t); got != mcp.CodeInvalidParams {
		t.Errorf("expected invalid params %d, got %d", mcp.CodeInvalidParams, got)
	}
	if !strings.Contains(res.err["message"].(string), "Unknown tool") {
		t.Errorf("unexpected message %v", res.err["message"])
	}
}

func TestMCPCallMissingToolName(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{}}`)
	if got := res.code(t); got != mcp.CodeInvalidParams {
		t.Errorf("expected invalid params, got %d", got)
	}
}

func TestMCPCallToolErrorReturn(t *testing.T) {
	a := newMCPTestApp(t)
	RegisterMCPTool(MCPTool{Name: "fail", Handler: func(c *MCPContext) any {
		return fmt.Errorf("invoice 42 is already paid")
	}})

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"fail"}}`)
	if res.result["isError"] != true {
		t.Fatalf("an error return must set isError, got %s", res.body)
	}
	text := res.result["content"].([]any)[0].(map[string]any)["text"]
	if text != "invoice 42 is already paid" {
		t.Errorf("unexpected message %v", text)
	}
}

func TestMCPCallToolPanicIsContained(t *testing.T) {
	a := newMCPTestApp(t)
	RegisterMCPTool(MCPTool{Name: "boom", Handler: func(c *MCPContext) any {
		panic("nil map write")
	}})

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"boom"}}`)
	if res.status != nethttp.StatusOK {
		t.Fatalf("a panicking tool must not break the response, got %d", res.status)
	}
	if res.result["isError"] != true {
		t.Fatalf("expected a contained tool error, got %s", res.body)
	}
	// The panic value itself must not leak to the client.
	text := res.result["content"].([]any)[0].(map[string]any)["text"].(string)
	if strings.Contains(text, "nil map write") {
		t.Errorf("panic detail must not be exposed, got %q", text)
	}
}

func TestMCPCallToolVerbatimResult(t *testing.T) {
	a := newMCPTestApp(t)
	RegisterMCPTool(MCPTool{Name: "raw", Handler: func(c *MCPContext) any {
		return &mcp.CallToolResult{Content: []mcp.Content{
			{Type: "image", Data: "AAAA", MimeType: "image/png"},
		}}
	}})

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"raw"}}`)
	block := res.result["content"].([]any)[0].(map[string]any)
	if block["type"] != "image" || block["mimeType"] != "image/png" {
		t.Errorf("a *mcp.CallToolResult must pass through untouched, got %v", block)
	}
}

// --- MCPContext ------------------------------------------------------------

func TestMCPContextExposesRequestAndCall(t *testing.T) {
	a := newMCPTestApp(t)

	var (
		seenHeader  string
		seenClient  string
		seenVersion string
		seenID      any
		seenArg     string
		ctxIsLive   bool
	)
	RegisterMCPTool(MCPTool{Name: "inspect", Handler: func(c *MCPContext) any {
		seenHeader = c.Header("X-Tenant") // promoted from *Request
		seenClient = c.Client.Name
		seenVersion = c.Version
		seenID = c.RequestID
		seenArg = c.Arg("filter.from").String() // dotted path into arguments
		ctxIsLive = c.Ctx != nil
		c.Progress(50, "halfway") // must be a safe no-op
		return "done"
	}})

	res := call(t, a, `{"jsonrpc":"2.0","id":"req-9","method":"tools/call","params":{
		"name":"inspect",
		"arguments":{"filter":{"from":"2026-01-01"}},
		"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28",
		         "io.modelcontextprotocol/clientInfo":{"name":"claude-code","version":"1.0"}}}}`,
		"X-Tenant", "acme", mcp.HeaderProtocolVersion, mcp.Version20260728)

	if res.err != nil {
		t.Fatalf("unexpected error %v", res.err)
	}
	if seenHeader != "acme" {
		t.Errorf("expected the request header to be reachable, got %q", seenHeader)
	}
	if seenClient != "claude-code" {
		t.Errorf("expected the client identity, got %q", seenClient)
	}
	if seenVersion != mcp.Version20260728 {
		t.Errorf("expected the declared version, got %q", seenVersion)
	}
	if seenID != "req-9" {
		t.Errorf("expected the JSON-RPC id, got %v", seenID)
	}
	if seenArg != "2026-01-01" {
		t.Errorf("expected Arg to resolve a dotted path, got %q", seenArg)
	}
	if !ctxIsLive {
		t.Error("Ctx must never be nil")
	}
}

func TestMCPContextArgOnMissingPath(t *testing.T) {
	a := newMCPTestApp(t)

	var empty bool
	RegisterMCPTool(MCPTool{Name: "probe", Handler: func(c *MCPContext) any {
		empty = c.Arg("nope.missing").IsNil()
		return "ok"
	}})

	call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"probe"}}`)
	if !empty {
		t.Error("Arg on a missing path should be nil, not panic")
	}
}

// --- auth ------------------------------------------------------------------

type permUser struct {
	DefaultUserInterface
	granted map[string]bool
}

func (u permUser) HasPermission(permission string) bool { return u.granted[permission] }
func (u permUser) Anonymous() bool                      { return false }
func (u permUser) FromRequest(r *Request) UserInterface { return u }

// withUser installs a UserInterface for the duration of a test.
func withUser(t *testing.T, granted ...string) {
	t.Helper()
	previous := UserInterfaceInstance
	t.Cleanup(func() { UserInterfaceInstance = previous })

	set := map[string]bool{}
	for _, p := range granted {
		set[p] = true
	}
	SetUserInterface(permUser{granted: set})
}

func registerPermissionedTools(t *testing.T) {
	t.Helper()
	RegisterMCPTool(
		MCPTool{Name: "public", Handler: func(c *MCPContext) any { return "public" }},
		MCPTool{Name: "secret", Permission: "invoice.read", Handler: func(c *MCPContext) any { return "secret" }},
	)
}

func TestMCPToolsListHidesUnpermittedTools(t *testing.T) {
	a := newMCPTestApp(t)
	registerPermissionedTools(t)
	withUser(t) // no permissions granted

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	tools := res.result["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("expected only the public tool, got %d", len(tools))
	}
	if tools[0].(map[string]any)["name"] != "public" {
		t.Errorf("the permissioned tool must be hidden, got %v", tools[0])
	}
}

func TestMCPToolsListShowsPermittedTools(t *testing.T) {
	a := newMCPTestApp(t)
	registerPermissionedTools(t)
	withUser(t, "invoice.read")

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	if len(res.result["tools"].([]any)) != 2 {
		t.Errorf("expected both tools, got %s", res.body)
	}
}

func TestMCPCallHiddenToolReportsUnknown(t *testing.T) {
	a := newMCPTestApp(t)
	registerPermissionedTools(t)
	withUser(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"secret"}}`)
	// A tool the caller cannot use must look absent, so the endpoint does not
	// disclose what it hides.
	if !strings.Contains(res.err["message"].(string), "Unknown tool") {
		t.Errorf("expected the tool to look absent, got %v", res.err)
	}
}

func TestMCPCallPermittedTool(t *testing.T) {
	a := newMCPTestApp(t)
	registerPermissionedTools(t)
	withUser(t, "invoice.read")

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"secret"}}`)
	if res.err != nil {
		t.Fatalf("expected success, got %v", res.err)
	}
	if res.result["content"].([]any)[0].(map[string]any)["text"] != "secret" {
		t.Errorf("unexpected result %s", res.body)
	}
}

func TestMCPBearerTokenIsEnforced(t *testing.T) {
	a := newMCPTestApp(t)
	mcpConfig.Token = "s3cret"
	registerEchoTool(t)

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`

	if res := call(t, a, body); res.status != nethttp.StatusUnauthorized {
		t.Errorf("expected 401 without a token, got %d", res.status)
	}
	if res := call(t, a, body, "Authorization", "Bearer wrong"); res.status != nethttp.StatusUnauthorized {
		t.Errorf("expected 401 with a wrong token, got %d", res.status)
	}
	if res := call(t, a, body, "Authorization", "Basic s3cret"); res.status != nethttp.StatusUnauthorized {
		t.Errorf("expected 401 for a non-bearer scheme, got %d", res.status)
	}
	res := call(t, a, body, "Authorization", "Bearer s3cret")
	if res.status != nethttp.StatusOK {
		t.Errorf("expected 200 with the right token, got %d body %s", res.status, res.body)
	}
	// The scheme is case-insensitive per RFC 7235.
	if res := call(t, a, body, "Authorization", "bearer s3cret"); res.status != nethttp.StatusOK {
		t.Errorf("the bearer scheme should be case-insensitive, got %d", res.status)
	}
}

func TestMCPNoTokenConfiguredSkipsTokenCheck(t *testing.T) {
	a := newMCPTestApp(t)

	res := call(t, a, `{"jsonrpc":"2.0","id":1,"method":"ping"}`)
	if res.status != nethttp.StatusOK {
		t.Errorf("expected 200 when no token is configured, got %d", res.status)
	}
}

func TestMCPOriginIsValidated(t *testing.T) {
	a := newMCPTestApp(t)
	body := `{"jsonrpc":"2.0","id":1,"method":"ping"}`

	// No allow list: a browser origin is refused, a native client that sends
	// no Origin at all is allowed.
	if res := call(t, a, body, "Origin", "https://evil.example"); res.status != nethttp.StatusForbidden {
		t.Errorf("expected 403 for an unlisted origin, got %d", res.status)
	}
	if res := call(t, a, body); res.status != nethttp.StatusOK {
		t.Errorf("a request without Origin must pass, got %d", res.status)
	}

	mcpConfig.AllowedOrigins = "https://app.example, https://admin.example"
	if res := call(t, a, body, "Origin", "https://app.example"); res.status != nethttp.StatusOK {
		t.Errorf("expected 200 for a listed origin, got %d", res.status)
	}
	if res := call(t, a, body, "Origin", "https://ADMIN.example"); res.status != nethttp.StatusOK {
		t.Errorf("origin comparison should be case-insensitive, got %d", res.status)
	}
	if res := call(t, a, body, "Origin", "https://evil.example"); res.status != nethttp.StatusForbidden {
		t.Errorf("expected 403 for an unlisted origin, got %d", res.status)
	}

	mcpConfig.AllowedOrigins = "*"
	if res := call(t, a, body, "Origin", "https://anything.example"); res.status != nethttp.StatusOK {
		t.Errorf("* should allow any origin, got %d", res.status)
	}
}

// --- wiring ----------------------------------------------------------------

func TestRegisterMCPEndpointsRespectsEnabledFlag(t *testing.T) {
	a := newTestApp()

	previous := mcpConfig
	t.Cleanup(func() { mcpConfig = previous })

	mcpConfig = MCPConfig{Enabled: false, Path: "/mcp-disabled"}
	registerMCPEndpoints()

	res := do(t, a, httptest.NewRequest("POST", "/mcp-disabled", strings.NewReader("{}")))
	if res.status != nethttp.StatusNotFound {
		t.Errorf("a disabled endpoint must not be routed, got %d", res.status)
	}
}

func TestRegisterMCPEndpointsMountsConfiguredPath(t *testing.T) {
	a := newTestApp()

	previous := mcpConfig
	t.Cleanup(func() { mcpConfig = previous })

	mcpConfig = MCPConfig{Enabled: true, Path: "/api/mcp", Name: "t", Version: "1"}
	registerMCPEndpoints()

	req := httptest.NewRequest("POST", "/api/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	req.Header.Set("Content-Type", "application/json")
	if res := do(t, a, req); res.status != nethttp.StatusOK {
		t.Errorf("expected the endpoint at the configured path, got %d", res.status)
	}
}
