package mcp

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Schema is a JSON Schema (draft 2020-12) document. Only the keywords this
// package can derive from Go types and evo validation tags are modelled.
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Enum        []any              `json:"enum,omitempty"`
	Format      string             `json:"format,omitempty"`
	Pattern     string             `json:"pattern,omitempty"`
	Default     any                `json:"default,omitempty"`

	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`

	MinLength *int `json:"minLength,omitempty"`
	MaxLength *int `json:"maxLength,omitempty"`
	MinItems  *int `json:"minItems,omitempty"`
	MaxItems  *int `json:"maxItems,omitempty"`

	UniqueItems          bool  `json:"uniqueItems,omitempty"`
	AdditionalProperties *bool `json:"additionalProperties,omitempty"`
}

var timeType = reflect.TypeOf(time.Time{})

// EmptyObjectSchema returns the schema a tool without parameters must expose:
// an object that accepts nothing else.
func EmptyObjectSchema() *Schema {
	deny := false
	return &Schema{Type: "object", AdditionalProperties: &deny}
}

// GenerateSchema derives a JSON Schema from a Go value. The value is only
// inspected for its type, so a zero value such as GetInvoiceInput{} is enough.
//
// Field names come from the `json` tag and fall back to the Go field name.
// Descriptions come from the `description` tag, defaults from the `default`
// tag, and constraints from the `validation` tag — the same tags the rest of
// the framework already uses.
//
// A nil input, or any input that is not a struct, yields EmptyObjectSchema.
func GenerateSchema(v any) *Schema {
	if v == nil {
		return EmptyObjectSchema()
	}
	t := reflect.TypeOf(v)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct || t == timeType {
		return EmptyObjectSchema()
	}
	return structSchema(t, map[reflect.Type]bool{})
}

// structSchema builds the object schema of a struct type. seen guards against
// infinite recursion on self-referencing types.
func structSchema(t reflect.Type, seen map[reflect.Type]bool) *Schema {
	if seen[t] {
		return &Schema{Type: "object"}
	}
	seen[t] = true
	defer delete(seen, t)

	schema := &Schema{Type: "object", Properties: map[string]*Schema{}}
	collectFields(t, schema, seen)
	if len(schema.Properties) == 0 {
		schema.Properties = nil
		deny := false
		schema.AdditionalProperties = &deny
	}
	return schema
}

// collectFields adds every exported field of t to schema, flattening embedded
// structs that carry no json tag of their own.
func collectFields(t reflect.Type, schema *Schema, seen map[reflect.Type]bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Unexported fields are skipped, except embedded ones: an embedded
		// unexported struct type still promotes its exported fields, the same
		// way encoding/json treats it.
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		jsonTag := field.Tag.Get("json")
		name, _, _ := strings.Cut(jsonTag, ",")
		if name == "-" {
			continue
		}

		// Embedded struct with no explicit name is flattened into the parent,
		// matching how encoding/json treats it.
		if field.Anonymous && name == "" {
			embedded := field.Type
			for embedded.Kind() == reflect.Ptr {
				embedded = embedded.Elem()
			}
			if embedded.Kind() == reflect.Struct && embedded != timeType {
				collectFields(embedded, schema, seen)
				continue
			}
		}

		if name == "" {
			name = field.Name
		}

		rules := parseRules(field.Tag.Get("validation"))
		prop := fieldSchema(field.Type, seen)
		prop.Description = field.Tag.Get("description")
		applyRules(prop, rules)

		// The default is coerced after the rules have run, because a rule such
		// as +int can still change the property's type.
		if def := field.Tag.Get("default"); def != "" {
			prop.Default = coerce(def, prop.Type)
		}

		schema.Properties[name] = prop
		if rules["required"] {
			schema.Required = append(schema.Required, name)
		}
	}
}

// fieldSchema maps a Go type onto its JSON Schema counterpart.
func fieldSchema(t reflect.Type, seen map[reflect.Type]bool) *Schema {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t == timeType {
		return &Schema{Type: "string", Format: "date-time"}
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Slice, reflect.Array:
		// []byte is conventionally a base64 string in JSON.
		if t.Elem().Kind() == reflect.Uint8 && t.Kind() == reflect.Slice {
			return &Schema{Type: "string"}
		}
		return &Schema{Type: "array", Items: fieldSchema(t.Elem(), seen)}
	case reflect.Map:
		return &Schema{Type: "object"}
	case reflect.Struct:
		return structSchema(t, seen)
	case reflect.Interface:
		return &Schema{} // any type
	default:
		return &Schema{Type: "string"}
	}
}

// parseRules splits an evo `validation` tag into a set of rules. Commas inside
// parentheses do not separate rules, mirroring the validation package's own
// parser so that rules like in(a,b,c) survive intact.
func parseRules(tag string) map[string]bool {
	rules := map[string]bool{}
	if tag == "" {
		return rules
	}
	var (
		current strings.Builder
		depth   int
	)
	flush := func() {
		if v := strings.TrimSpace(current.String()); v != "" {
			rules[v] = true
		}
		current.Reset()
	}
	for i := 0; i < len(tag); i++ {
		c := tag[i]
		switch {
		case c == '\\' && i+1 < len(tag):
			i++
			current.WriteByte(tag[i])
		case c == '(':
			depth++
			current.WriteByte(c)
		case c == ')':
			if depth > 0 {
				depth--
			}
			current.WriteByte(c)
		case c == ',' && depth == 0:
			flush()
		default:
			current.WriteByte(c)
		}
	}
	flush()
	return rules
}

// simple maps parameterless validators onto a JSON Schema `format` or
// `pattern`. Keys are lower case.
var simpleFormats = map[string]string{
	"email":  "email",
	"url":    "uri",
	"uuid":   "uuid",
	"date":   "date-time",
	"time":   "date-time",
	"ipv4":   "ipv4",
	"ip":     "ipv4",
	"ipv6":   "ipv6",
	"domain": "hostname",
}

var simplePatterns = map[string]string{
	"alpha":        `^[a-zA-Z]+$`,
	"alphanumeric": `^[a-zA-Z0-9]+$`,
	"digit":        `^[0-9]+$`,
	"slug":         `^[a-z0-9]+(?:-[a-z0-9]+)*$`,
	"hex":          `^[0-9a-fA-F]+$`,
	"lowercase":    `^[^A-Z]*$`,
	"uppercase":    `^[^a-z]*$`,
}

// applyRules folds evo validation rules into JSON Schema keywords. Rules with
// no schema equivalent are ignored; validation still enforces them at call
// time, the schema only advertises what a client can check up front.
func applyRules(s *Schema, rules map[string]bool) {
	// Rules live in a map, so they arrive in no particular order. Rules that
	// change the property's type are applied first, because every other rule
	// reads that type to decide how to interpret itself.
	for rule := range rules {
		switch strings.ToLower(rule) {
		case "int", "+int", "-int":
			s.Type = "integer"
		case "float", "+float", "-float":
			s.Type = "number"
		}
	}
	numeric := s.Type == "integer" || s.Type == "number"

	for rule := range rules {
		lower := strings.ToLower(rule)

		if format, ok := simpleFormats[lower]; ok && s.Type == "string" {
			s.Format = format
			continue
		}
		if pattern, ok := simplePatterns[lower]; ok && s.Type == "string" {
			s.Pattern = pattern
			continue
		}

		switch lower {
		case "required", "int", "float":
			continue
		case "+int", "+float":
			s.Minimum = f64(0)
			continue
		case "-int", "-float":
			s.Maximum = f64(0)
			continue
		case "unique_items", "unique-items", "uniqueitems":
			s.UniqueItems = true
			continue
		}

		// Arguments are read from the original rule, not the lower-cased copy:
		// a pattern or an enum value is case sensitive.
		if inner, ok := argOf(rule, "regex"); ok {
			s.Pattern = inner
			continue
		}
		if inner, ok := argOf(rule, "in"); ok {
			// Enum members must carry the property's own type: ["1","2"] can
			// never be satisfied by an integer property.
			for _, member := range splitTrim(inner) {
				s.Enum = append(s.Enum, coerce(member, s.Type))
			}
			continue
		}
		if inner, ok := argOf(rule, "min_items", "min-items", "minitems"); ok {
			if n, err := strconv.Atoi(strings.TrimSpace(inner)); err == nil {
				s.MinItems = &n
			}
			continue
		}
		if inner, ok := argOf(rule, "max_items", "max-items", "maxitems"); ok {
			if n, err := strconv.Atoi(strings.TrimSpace(inner)); err == nil {
				s.MaxItems = &n
			}
			continue
		}

		// len<N / len>=N / len==N ... apply to strings and arrays.
		if op, n, ok := comparison(lower, "len"); ok {
			applyLength(s, op, n)
			continue
		}
		// >N / >=N / ==N ... apply to numbers only.
		if op, n, ok := comparison(lower, ""); ok && numeric {
			applyBound(s, op, float64(n))
			continue
		}
	}
}

// applyLength turns a len comparison into minLength/maxLength for strings or
// minItems/maxItems for arrays.
func applyLength(s *Schema, op string, n int) {
	setMin, setMax := func(v int) { s.MinLength = &v }, func(v int) { s.MaxLength = &v }
	if s.Type == "array" {
		setMin, setMax = func(v int) { s.MinItems = &v }, func(v int) { s.MaxItems = &v }
	}

	switch op {
	case ">":
		setMin(n + 1)
	case ">=":
		setMin(n)
	case "<":
		if n > 0 {
			setMax(n - 1)
		} else {
			setMax(0)
		}
	case "<=":
		setMax(n)
	case "==", "=":
		setMin(n)
		setMax(n)
	}
}

// applyBound turns a numeric comparison into minimum/maximum bounds.
func applyBound(s *Schema, op string, n float64) {
	switch op {
	case ">":
		s.ExclusiveMinimum = f64(n)
	case ">=":
		s.Minimum = f64(n)
	case "<":
		s.ExclusiveMaximum = f64(n)
	case "<=":
		s.Maximum = f64(n)
	case "==", "=":
		s.Minimum = f64(n)
		s.Maximum = f64(n)
	}
}

// comparison parses "<prefix><op><integer>" and reports the operator and value.
// prefix is "len" for length rules and "" for plain numeric rules.
func comparison(rule, prefix string) (string, int, bool) {
	if !strings.HasPrefix(rule, prefix) {
		return "", 0, false
	}
	rest := rule[len(prefix):]
	for _, op := range []string{">=", "<=", "==", "!=", "<>", ">", "<", "="} {
		if strings.HasPrefix(rest, op) {
			if op == "!=" || op == "<>" {
				return "", 0, false // no JSON Schema equivalent
			}
			n, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(rest[len(op):]), "+"))
			if err != nil {
				return "", 0, false
			}
			return op, n, true
		}
	}
	return "", 0, false
}

// argOf reports the argument of rule when it has the form "name(argument)" for
// any of the accepted names. The name is matched case-insensitively; the
// argument is returned with its original casing intact.
func argOf(rule string, names ...string) (string, bool) {
	if !strings.HasSuffix(rule, ")") {
		return "", false
	}
	for _, name := range names {
		if len(rule) > len(name)+1 &&
			strings.EqualFold(rule[:len(name)], name) &&
			rule[len(name)] == '(' {
			return rule[len(name)+1 : len(rule)-1], true
		}
	}
	return "", false
}

// coerce converts a tag-supplied literal into the Go type that matches the
// schema type it will sit next to, so that an enum member or a default is not
// advertised as a string on a numeric property. A value that does not parse is
// left as a string rather than dropped.
func coerce(value, schemaType string) any {
	value = strings.TrimSpace(value)
	switch schemaType {
	case "integer":
		if n, err := strconv.ParseInt(value, 10, 64); err == nil {
			return n
		}
	case "number":
		if n, err := strconv.ParseFloat(value, 64); err == nil {
			return n
		}
	case "boolean":
		switch strings.ToLower(value) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	}
	return value
}

func splitTrim(in string) []string {
	parts := strings.Split(in, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func f64(v float64) *float64 { return &v }
