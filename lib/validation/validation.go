package validation

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/generic"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
)

// CustomValidators allows registration of custom validators
var CustomValidators = make(map[*regexp.Regexp]func(match []string, value *generic.Value) error)
var customValidatorsMutex sync.RWMutex

// CustomDBValidators allows registration of custom DB validators
var CustomDBValidators = make(map[*regexp.Regexp]func(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error)
var customDBValidatorsMutex sync.RWMutex

// RegisterValidator registers a custom validator
func RegisterValidator(pattern string, fn func(match []string, value *generic.Value) error) {
	customValidatorsMutex.Lock()
	defer customValidatorsMutex.Unlock()
	CustomValidators[regexp.MustCompile(pattern)] = fn
}

// RegisterDBValidator registers a custom database validator
func RegisterDBValidator(pattern string, fn func(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error) {
	customDBValidatorsMutex.Lock()
	defer customDBValidatorsMutex.Unlock()
	CustomDBValidators[regexp.MustCompile(pattern)] = fn
}

// Struct validates a struct with all fields
// Deprecated: Use StructWithContext for better context support
func Struct(input interface{}, fields ...string) []error {
	return StructWithContext(context.Background(), input, fields...)
}

// StructWithContext validates a struct with context support
func StructWithContext(ctx context.Context, input interface{}, fields ...string) []error {
	var errors []error

	ref := reflect.ValueOf(input)
	for ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	if ref.Kind() != reflect.Struct {
		return nil
	}

	var g = generic.Parse(input)

	dbInstance := db.GetInstance()
	if dbInstance == nil {
		return []error{fmt.Errorf("database not initialized - only non-DB validators available")}
	}

	var stmt = dbInstance.Model(input).Statement
	if err := stmt.Parse(input); err != nil {
		return []error{fmt.Errorf("failed to parse model: %w", err)}
	}

	for idx, _ := range stmt.Schema.Fields {
		field := stmt.Schema.Fields[idx]
		if field.Tag.Get("validation") != "" {
			var err = validateFieldWithContext(ctx, &g, field, stmt)
			if err != nil {
				errors = append(errors, err)
			}
		}

	}

	return errors
}

// StructNonZeroFields validates only non-zero fields
// Deprecated: Use StructNonZeroFieldsWithContext for better context support
func StructNonZeroFields(input interface{}, fields ...string) []error {
	return StructNonZeroFieldsWithContext(context.Background(), input, fields...)
}

// StructNonZeroFieldsWithContext validates only non-zero fields with context support
func StructNonZeroFieldsWithContext(ctx context.Context, input interface{}, fields ...string) []error {
	var errors []error

	ref := reflect.ValueOf(input)
	for ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	if ref.Kind() != reflect.Struct {
		return nil
	}

	var g = generic.Parse(input)

	dbInstance := db.GetInstance()
	if dbInstance == nil {
		return []error{fmt.Errorf("database not initialized - only non-DB validators available")}
	}

	var stmt = dbInstance.Model(input).Statement
	if err := stmt.Parse(input); err != nil {
		return []error{fmt.Errorf("failed to parse model: %w", err)}
	}

	ref = reflect.ValueOf(input)
	for idx, _ := range stmt.Schema.Fields {
		field := stmt.Schema.Fields[idx]
		_, zero := field.ValueOf(ctx, ref)

		if !zero && field.Tag.Get("validation") != "" {
			var err = validateFieldWithContext(ctx, &g, field, stmt)
			if err != nil {
				errors = append(errors, err)
			}
		}

	}

	return errors
}

func validateField(g *generic.Value, field *schema.Field) error {
	return validateFieldWithContext(context.Background(), g, field, nil)
}

func validateFieldWithContext(ctx context.Context, g *generic.Value, field *schema.Field, stmt *gorm.Statement) error {
	if stmt == nil {
		dbInstance := db.GetInstance()
		if dbInstance == nil {
			// Skip DB validators if database is not available
			return validateFieldWithoutDB(g, field)
		}
		var s = dbInstance.Model(g.Input).Statement
		var err = s.Parse(g.Input)
		if err != nil {
			return err
		}
		stmt = s
	}

	var value = g.Prop(field.Name)
	validators := parseValidators(field.Tag.Get("validation"))
	for _, validator := range validators {
		var found = false

		// Check custom DB validators first
		customDBValidatorsMutex.RLock()
		for r, fn := range CustomDBValidators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				err := fn(match, &value, stmt, stmt.Schema.FieldsByName[field.Name])
				customDBValidatorsMutex.RUnlock()
				tag := field.Tag.Get("json")
				if tag == "" {
					tag = field.Name
				}
				if err != nil {
					return fmt.Errorf("%s %s", tag, err)
				}
				goto nextValidator
			}
		}
		customDBValidatorsMutex.RUnlock()

		// Check built-in DB validators
		for r, fn := range DBValidators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				err := fn(match, &value, stmt, stmt.Schema.FieldsByName[field.Name])
				tag := field.Tag.Get("json")
				if tag == "" {
					tag = field.Name
				}
				if err != nil {
					return fmt.Errorf("%s %s", tag, err)
				}
				goto nextValidator
			}
		}

		// Check custom validators
		customValidatorsMutex.RLock()
		for r, fn := range CustomValidators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				var err = fn(match, &value)
				customValidatorsMutex.RUnlock()
				tag := field.Tag.Get("json")
				if tag == "" {
					tag = field.Name
				}
				if err != nil {
					return fmt.Errorf("%s %s", tag, err)
				}
				goto nextValidator
			}
		}
		customValidatorsMutex.RUnlock()

		// Check built-in validators
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
				goto nextValidator
			}
		}

		if !found {
			return fmt.Errorf("validator %s not found for %s.%s", validator, g.IndirectType().Name(), field.Name)
		}

		nextValidator:
	}
	return nil
}

// validateFieldWithoutDB validates a field without database validators
func validateFieldWithoutDB(g *generic.Value, field *schema.Field) error {
	var value = g.Prop(field.Name)
	validators := parseValidators(field.Tag.Get("validation"))
	for _, validator := range validators {
		var found = false

		// Skip DB validators (unique, fk, enum, before, after, confirmed, same, different, gt_field, etc.)
		// These will be silently skipped when database is not available

		// Check custom validators
		customValidatorsMutex.RLock()
		for r, fn := range CustomValidators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				var err = fn(match, &value)
				customValidatorsMutex.RUnlock()
				tag := field.Tag.Get("json")
				if tag == "" {
					tag = field.Name
				}
				if err != nil {
					return fmt.Errorf("%s %s", tag, err)
				}
				goto nextValidatorNoDB
			}
		}
		customValidatorsMutex.RUnlock()

		// Check built-in validators
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
				goto nextValidatorNoDB
			}
		}

		// Don't error on DB validators when DB is not available
		for r := range DBValidators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				goto nextValidatorNoDB
			}
		}
		for r := range CustomDBValidators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				goto nextValidatorNoDB
			}
		}

		if !found {
			return fmt.Errorf("validator %s not found for %s.%s", validator, g.IndirectType().Name(), field.Name)
		}

		nextValidatorNoDB:
	}
	return nil
}

func parseValidators(s string) []string {
	var result []string
	var buffer = ""
	var lastChar rune
	var parenDepth = 0

	for _, c := range s {
		if c == '(' {
			parenDepth++
			buffer += string(c)
		} else if c == ')' {
			parenDepth--
			buffer += string(c)
		} else if c == ',' && lastChar != '\\' && parenDepth == 0 {
			// Only split on comma if we're not inside parentheses
			if len(buffer) > 0 {
				result = append(result, buffer)
				buffer = ""
			}
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

// Value validates a single value
// Deprecated: Use ValueWithContext for better context support
func Value(input interface{}, validation string) error {
	return ValueWithContext(context.Background(), input, validation)
}

// ValueWithContext validates a single value with context support
func ValueWithContext(ctx context.Context, input interface{}, validation string) error {
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

		// Check custom validators first
		customValidatorsMutex.RLock()
		for r, fn := range CustomValidators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				var err = fn(match, g)
				customValidatorsMutex.RUnlock()
				if err != nil {
					return err
				}
				goto nextValueValidator
			}
		}
		customValidatorsMutex.RUnlock()

		// Check built-in validators
		for r, fn := range Validators {
			if match := r.FindStringSubmatch(validator); len(match) > 0 {
				found = true
				var err = fn(match, g)
				if err != nil {
					return err
				}
				goto nextValueValidator
			}
		}

		if !found {
			return fmt.Errorf("validator %s not found", validator)
		}

		nextValueValidator:
	}
	return nil
}
