package validation

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/generic"
	"reflect"
)

func Struct(input interface{}, fields ...string) []error {
	var errors []error
	var g = generic.Parse(input)
	if len(fields) == 0 {
		fields = g.FieldNames()
	}

	for _, field := range g.Props() {
		for _, f := range fields {
			if field.Name == f {
				if field.Tag.Get("validation") != "" {

					var err = validateField(&g, &field)
					if err != nil {
						errors = append(errors, err)
					}
				}
				break
			}
		}
	}

	return errors
}

func validateField(g *generic.Value, field *reflect.StructField) error {
	var value = g.Prop(field.Name)
	validators := parseValidators(field.Tag.Get("validation"))
	for _, validator := range validators {
		var found = false
		for r, fn := range Validators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				var err = fn(match, &value)
				tag := field.Tag.Get("json")
				if tag == "" {
					tag = field.Name
				}
				if err != nil {
					return fmt.Errorf("%s %s", tag, err)
				}
			}
		}
		if !found {
			return fmt.Errorf("validator %s not found for %s.%s", validator, g.IndirectType().Name(), field.Name)
		}
	}
	return nil
}

func parseValidators(s string) []string {
	var result []string
	var buffer = ""
	var lastChar rune
	for _, c := range s {
		if c == ',' && lastChar != '\\' {
			result = append(result, buffer)
			buffer = ""
		} else {
			buffer += string(c)
		}
		lastChar = c
	}
	if len(buffer) > 0 {
		result = append(result, buffer)
	}
	return result
}

func Value(input interface{}, validation string) error {
	var g *generic.Value
	if v, ok := input.(*generic.Value); ok {
		g = v
	} else {
		var parsed = generic.Parse(input)
		g = &parsed
	}

	validators := parseValidators(validation)
	for _, validator := range validators {
		var found = false
		for r, fn := range Validators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				var err = fn(match, g)
				if err != nil {
					return err
				}
			}
		}
		if !found {
			return fmt.Errorf("validator %s not found", validator)
		}

	}
	return nil
}
