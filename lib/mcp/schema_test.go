package mcp

import (
	"testing"
	"time"

	"github.com/getevo/json"
)

func TestGenerateSchemaNilYieldsEmptyObject(t *testing.T) {
	s := GenerateSchema(nil)
	if s.Type != "object" {
		t.Fatalf("expected object, got %q", s.Type)
	}
	if s.AdditionalProperties == nil || *s.AdditionalProperties {
		t.Fatal("expected additionalProperties:false for a tool with no arguments")
	}
	if len(s.Properties) != 0 {
		t.Fatalf("expected no properties, got %d", len(s.Properties))
	}
}

func TestGenerateSchemaNonStructYieldsEmptyObject(t *testing.T) {
	for _, input := range []any{"str", 42, []string{"a"}, time.Time{}} {
		s := GenerateSchema(input)
		if s.Type != "object" || s.Properties != nil {
			t.Fatalf("expected empty object schema for %T, got %+v", input, s)
		}
	}
}

func TestGenerateSchemaTypes(t *testing.T) {
	type nested struct {
		City string `json:"city"`
	}
	type input struct {
		Name    string    `json:"name"`
		Age     int       `json:"age"`
		Score   float64   `json:"score"`
		Active  bool      `json:"active"`
		Tags    []string  `json:"tags"`
		Address nested    `json:"address"`
		When    time.Time `json:"when"`
		Blob    []byte    `json:"blob"`
		Extra   map[string]string
		Any     any    `json:"any"`
		Ptr     *int   `json:"ptr"`
		Skipped string `json:"-"`
		private string
	}

	s := GenerateSchema(input{})

	expected := map[string]string{
		"name": "string", "age": "integer", "score": "number", "active": "boolean",
		"tags": "array", "address": "object", "when": "string", "blob": "string",
		"Extra": "object", "ptr": "integer",
	}
	for field, want := range expected {
		prop, ok := s.Properties[field]
		if !ok {
			t.Fatalf("missing property %q", field)
		}
		if prop.Type != want {
			t.Errorf("%s: expected type %q, got %q", field, want, prop.Type)
		}
	}

	if _, ok := s.Properties["-"]; ok {
		t.Error(`json:"-" field must be skipped`)
	}
	if _, ok := s.Properties["Skipped"]; ok {
		t.Error(`json:"-" field must be skipped`)
	}
	if _, ok := s.Properties["private"]; ok {
		t.Error("unexported field must be skipped")
	}
	if s.Properties["any"].Type != "" {
		t.Error("interface field must have no type constraint")
	}
	if s.Properties["tags"].Items == nil || s.Properties["tags"].Items.Type != "string" {
		t.Error("expected tags.items.type == string")
	}
	if s.Properties["when"].Format != "date-time" {
		t.Errorf("expected time.Time to carry format date-time, got %q", s.Properties["when"].Format)
	}
	if s.Properties["address"].Properties["city"].Type != "string" {
		t.Error("expected nested struct properties to be walked")
	}
}

func TestGenerateSchemaFieldNaming(t *testing.T) {
	type input struct {
		Explicit string `json:"explicit_name"`
		WithOpts string `json:"with_opts,omitempty"`
		NoTag    string
	}
	s := GenerateSchema(input{})
	for _, name := range []string{"explicit_name", "with_opts", "NoTag"} {
		if _, ok := s.Properties[name]; !ok {
			t.Errorf("expected property %q, have %v", name, keys(s.Properties))
		}
	}
}

func TestGenerateSchemaEmbeddedIsFlattened(t *testing.T) {
	type base struct {
		ID string `json:"id" validation:"required"`
	}
	type named struct {
		Value string `json:"value"`
	}
	type input struct {
		base
		Named named  `json:"named"`
		Extra string `json:"extra"`
	}

	s := GenerateSchema(input{})
	if _, ok := s.Properties["id"]; !ok {
		t.Fatalf("embedded field must be flattened into the parent, have %v", keys(s.Properties))
	}
	if !contains(s.Required, "id") {
		t.Error("required on an embedded field must reach the parent required list")
	}
	if _, ok := s.Properties["named"]; !ok {
		t.Error("named struct field must stay nested")
	}
}

func TestGenerateSchemaRequired(t *testing.T) {
	type input struct {
		Must  string `json:"must" validation:"required"`
		Maybe string `json:"maybe"`
	}
	s := GenerateSchema(input{})
	if !contains(s.Required, "must") {
		t.Errorf("expected must to be required, got %v", s.Required)
	}
	if contains(s.Required, "maybe") {
		t.Errorf("maybe must not be required, got %v", s.Required)
	}
}

func TestGenerateSchemaDescriptionAndDefault(t *testing.T) {
	type input struct {
		Limit int     `json:"limit" description:"how many rows to return" default:"25"`
		Rate  float64 `json:"rate" default:"1.5"`
		On    bool    `json:"on" default:"true"`
		Label string  `json:"label" default:"none"`
		Odd   int     `json:"odd" default:"not-a-number"`
	}
	s := GenerateSchema(input{})

	if got := s.Properties["limit"].Description; got != "how many rows to return" {
		t.Errorf("unexpected description %q", got)
	}
	// A default must carry the property's own type. "25" on an integer
	// property is not a valid default.
	if got := s.Properties["limit"].Default; got != int64(25) {
		t.Errorf("expected int64(25), got %T(%v)", got, got)
	}
	if got := s.Properties["rate"].Default; got != 1.5 {
		t.Errorf("expected 1.5, got %T(%v)", got, got)
	}
	if got := s.Properties["on"].Default; got != true {
		t.Errorf("expected true, got %T(%v)", got, got)
	}
	if got := s.Properties["label"].Default; got != "none" {
		t.Errorf("expected \"none\", got %T(%v)", got, got)
	}
	// An unparseable default falls back to the literal rather than vanishing.
	if got := s.Properties["odd"].Default; got != "not-a-number" {
		t.Errorf("expected the raw literal, got %T(%v)", got, got)
	}
}

func TestGenerateSchemaTypedEnum(t *testing.T) {
	type input struct {
		Level   int     `json:"level" validation:"in(1,2,3)"`
		Ratio   float64 `json:"ratio" validation:"in(0.5,1.5)"`
		Status  string  `json:"status" validation:"in(draft,paid)"`
		Coerced string  `json:"coerced" validation:"int,in(4,5)"`
	}
	s := GenerateSchema(input{})

	// An integer property must not advertise string enum members: no integer
	// could ever satisfy ["1","2","3"].
	for _, member := range s.Properties["level"].Enum {
		if _, ok := member.(int64); !ok {
			t.Errorf("level: expected int64 enum members, got %T(%v)", member, member)
		}
	}
	for _, member := range s.Properties["ratio"].Enum {
		if _, ok := member.(float64); !ok {
			t.Errorf("ratio: expected float64 enum members, got %T(%v)", member, member)
		}
	}
	for _, member := range s.Properties["status"].Enum {
		if _, ok := member.(string); !ok {
			t.Errorf("status: expected string enum members, got %T(%v)", member, member)
		}
	}
	// A type-changing rule must win regardless of map iteration order, so the
	// enum members follow the type the rule imposed.
	if s.Properties["coerced"].Type != "integer" {
		t.Fatalf("expected the int rule to force integer, got %q", s.Properties["coerced"].Type)
	}
	for _, member := range s.Properties["coerced"].Enum {
		if _, ok := member.(int64); !ok {
			t.Errorf("coerced: expected int64 enum members, got %T(%v)", member, member)
		}
	}
}

func TestGenerateSchemaTypeRuleOrderIsStable(t *testing.T) {
	type input struct {
		N string `json:"n" validation:"int,>=5,<=9"`
	}
	// Rules are held in a map, so run this enough times that a dependency on
	// iteration order would show up.
	for i := 0; i < 200; i++ {
		s := GenerateSchema(input{})
		prop := s.Properties["n"]
		if prop.Type != "integer" {
			t.Fatalf("iteration %d: expected integer, got %q", i, prop.Type)
		}
		if prop.Minimum == nil || *prop.Minimum != 5 {
			t.Fatalf("iteration %d: numeric bound was dropped: %+v", i, prop)
		}
		if prop.Maximum == nil || *prop.Maximum != 9 {
			t.Fatalf("iteration %d: numeric bound was dropped: %+v", i, prop)
		}
	}
}

func TestGenerateSchemaFormats(t *testing.T) {
	type input struct {
		Email string `json:"email" validation:"required,email"`
		URL   string `json:"url" validation:"url"`
		ID    string `json:"id" validation:"uuid"`
		When  string `json:"when" validation:"date"`
		Addr  string `json:"addr" validation:"ipv4"`
		Host  string `json:"host" validation:"domain"`
	}
	s := GenerateSchema(input{})
	expect := map[string]string{
		"email": "email", "url": "uri", "id": "uuid",
		"when": "date-time", "addr": "ipv4", "host": "hostname",
	}
	for field, want := range expect {
		if got := s.Properties[field].Format; got != want {
			t.Errorf("%s: expected format %q, got %q", field, want, got)
		}
	}
}

func TestGenerateSchemaLengthBounds(t *testing.T) {
	type input struct {
		A string `json:"a" validation:"len>3"`
		B string `json:"b" validation:"len>=3"`
		C string `json:"c" validation:"len<10"`
		D string `json:"d" validation:"len<=10"`
		E string `json:"e" validation:"len==5"`
	}
	s := GenerateSchema(input{})

	assertInt(t, "a.minLength", s.Properties["a"].MinLength, 4)
	assertInt(t, "b.minLength", s.Properties["b"].MinLength, 3)
	assertInt(t, "c.maxLength", s.Properties["c"].MaxLength, 9)
	assertInt(t, "d.maxLength", s.Properties["d"].MaxLength, 10)
	assertInt(t, "e.minLength", s.Properties["e"].MinLength, 5)
	assertInt(t, "e.maxLength", s.Properties["e"].MaxLength, 5)
}

func TestGenerateSchemaNumericBounds(t *testing.T) {
	type input struct {
		A int `json:"a" validation:">3"`
		B int `json:"b" validation:">=3"`
		C int `json:"c" validation:"<10"`
		D int `json:"d" validation:"<=10"`
		E int `json:"e" validation:"==7"`
		F int `json:"f" validation:"!=7"`
	}
	s := GenerateSchema(input{})

	assertFloat(t, "a.exclusiveMinimum", s.Properties["a"].ExclusiveMinimum, 3)
	assertFloat(t, "b.minimum", s.Properties["b"].Minimum, 3)
	assertFloat(t, "c.exclusiveMaximum", s.Properties["c"].ExclusiveMaximum, 10)
	assertFloat(t, "d.maximum", s.Properties["d"].Maximum, 10)
	assertFloat(t, "e.minimum", s.Properties["e"].Minimum, 7)
	assertFloat(t, "e.maximum", s.Properties["e"].Maximum, 7)

	// != has no JSON Schema equivalent and must be ignored, not misapplied.
	f := s.Properties["f"]
	if f.Minimum != nil || f.Maximum != nil || f.ExclusiveMinimum != nil || f.ExclusiveMaximum != nil {
		t.Error("!= must not produce a bound")
	}
}

func TestGenerateSchemaNumericRulesIgnoredOnStrings(t *testing.T) {
	type input struct {
		Name string `json:"name" validation:">3"`
	}
	s := GenerateSchema(input{})
	if s.Properties["name"].ExclusiveMinimum != nil {
		t.Error("a numeric bound must not be applied to a string field")
	}
}

func TestGenerateSchemaEnumAndPattern(t *testing.T) {
	type input struct {
		Status string `json:"status" validation:"in(draft,sent,paid)"`
		Code   string `json:"code" validation:"regex(^[A-Z]{3}$)"`
		Slug   string `json:"slug" validation:"alphanumeric"`
	}
	s := GenerateSchema(input{})

	enum := s.Properties["status"].Enum
	if len(enum) != 3 || enum[0] != "draft" || enum[2] != "paid" {
		t.Errorf("unexpected enum %v", enum)
	}
	if s.Properties["status"].Type != "string" {
		t.Errorf("expected a string property, got %q", s.Properties["status"].Type)
	}
	if got := s.Properties["code"].Pattern; got != "^[A-Z]{3}$" {
		t.Errorf("unexpected pattern %q", got)
	}
	if s.Properties["slug"].Pattern == "" {
		t.Error("alphanumeric should produce a pattern")
	}
}

func TestGenerateSchemaArrayRules(t *testing.T) {
	type input struct {
		Tags []string `json:"tags" validation:"min_items(1),max_items(5),unique_items"`
		Lens []int    `json:"lens" validation:"len>=2"`
	}
	s := GenerateSchema(input{})

	tags := s.Properties["tags"]
	assertInt(t, "tags.minItems", tags.MinItems, 1)
	assertInt(t, "tags.maxItems", tags.MaxItems, 5)
	if !tags.UniqueItems {
		t.Error("expected uniqueItems")
	}
	// A len rule on an array maps to item counts, not string lengths.
	assertInt(t, "lens.minItems", s.Properties["lens"].MinItems, 2)
	if s.Properties["lens"].MinLength != nil {
		t.Error("len on an array must not set minLength")
	}
}

func TestGenerateSchemaTypeOverrideRules(t *testing.T) {
	type input struct {
		Positive string `json:"positive" validation:"+int"`
		Ratio    string `json:"ratio" validation:"float"`
	}
	s := GenerateSchema(input{})

	if s.Properties["positive"].Type != "integer" {
		t.Errorf("+int must force integer, got %q", s.Properties["positive"].Type)
	}
	assertFloat(t, "positive.minimum", s.Properties["positive"].Minimum, 0)
	if s.Properties["ratio"].Type != "number" {
		t.Errorf("float must force number, got %q", s.Properties["ratio"].Type)
	}
}

func TestParseRulesKeepsParenthesisedCommas(t *testing.T) {
	rules := parseRules("required,in(a,b,c),regex(^x,y$),len<=10")
	for _, want := range []string{"required", "in(a,b,c)", "regex(^x,y$)", "len<=10"} {
		if !rules[want] {
			t.Errorf("expected rule %q, got %v", want, keys(rules))
		}
	}
	if len(rules) != 4 {
		t.Errorf("expected 4 rules, got %d: %v", len(rules), keys(rules))
	}
}

func TestGenerateSchemaSelfReferenceTerminates(t *testing.T) {
	type node struct {
		Name  string `json:"name"`
		Child *node  `json:"child"`
	}
	s := GenerateSchema(node{}) // must not recurse forever
	if s.Properties["child"].Type != "object" {
		t.Errorf("expected the recursive field to degrade to a bare object, got %+v", s.Properties["child"])
	}
}

func TestSchemaMarshalsAsValidJSONSchema(t *testing.T) {
	type input struct {
		Name string `json:"name" validation:"required,len<=64" description:"customer name"`
	}
	raw, err := json.Marshal(GenerateSchema(input{}))
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["type"] != "object" {
		t.Errorf("expected type object, got %v", decoded["type"])
	}
	props, ok := decoded["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties object, got %T", decoded["properties"])
	}
	name := props["name"].(map[string]any)
	if name["type"] != "string" || name["description"] != "customer name" {
		t.Errorf("unexpected name schema %v", name)
	}
	if name["maxLength"].(float64) != 64 {
		t.Errorf("unexpected maxLength %v", name["maxLength"])
	}
	// A zero-valued omitempty pointer must not leak a null keyword.
	if _, present := name["minLength"]; present {
		t.Error("minLength must be omitted when unset")
	}
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func assertInt(t *testing.T, label string, got *int, want int) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: expected %d, got nil", label, want)
		return
	}
	if *got != want {
		t.Errorf("%s: expected %d, got %d", label, want, *got)
	}
}

func assertFloat(t *testing.T, label string, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: expected %v, got nil", label, want)
		return
	}
	if *got != want {
		t.Errorf("%s: expected %v, got %v", label, want, *got)
	}
}
