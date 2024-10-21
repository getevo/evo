package validation

import (
	"context"
	"fmt"
	"github.com/getevo/evo/v2/lib/db"
	scm "github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/generic"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var DBValidators = map[*regexp.Regexp]func(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error{
	regexp.MustCompile("^unique$"):          uniqueValidator,
	regexp.MustCompile("^fk$"):              foreignKeyValidator,
	regexp.MustCompile("^enum$"):            enumValidator,
	regexp.MustCompile(`^before\((\w+)\)$`): beforeValidator,
	regexp.MustCompile(`^after\((\w+)\)$`):  afterValidator,
}

var Validators = map[*regexp.Regexp]func(match []string, value *generic.Value) error{
	regexp.MustCompile("^text$"):                           textValidator,
	regexp.MustCompile("^name$"):                           nameValidator,
	regexp.MustCompile("^alpha$"):                          alphaValidator,
	regexp.MustCompile("^latin$"):                          latinValidator,
	regexp.MustCompile("^name$"):                           nameValidator,
	regexp.MustCompile("^digit$"):                          digitValidator,
	regexp.MustCompile("^alphanumeric$"):                   alphaNumericValidator,
	regexp.MustCompile("^required$"):                       requiredValidator,
	regexp.MustCompile("^email$"):                          emailValidator,
	regexp.MustCompile(`^regex\((.*)\)$`):                  regexValidator,
	regexp.MustCompile(`^len(>|<|<=|>=|==|!=|<>|=)(\d+)$`): lenValidator,
	regexp.MustCompile(`^(>|<|<=|>=|==|!=|<>|=)(\d+)$`):    numericalValidator,
	regexp.MustCompile(`^([+\-]?)int$`):                    intValidator,
	regexp.MustCompile(`^([+\-]?)float$`):                  floatValidator,
	regexp.MustCompile(`^password\((.*)\)$`):               passwordValidator,
	regexp.MustCompile(`^domain$`):                         domainValidator,
	regexp.MustCompile(`^url$`):                            urlValidator,
	regexp.MustCompile(`^ip$`):                             ipValidator,
	regexp.MustCompile(`^date$`):                           dateValidator,
}

func latinValidator(match []string, value *generic.Value) error {
	var v = regexp.MustCompile(`^\p{L}*$`)
	matchString := v.MatchString(value.String())
	if !matchString {
		return fmt.Errorf("is not latin")
	}
	return nil
}

var enumRegex = regexp.MustCompile(`(?m)enum\(([^)]+)\)`)
var enumBodyRegex = regexp.MustCompile(`(?m)'([^']*)'`)

func enumValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	var v = value.String()
	if field.StructField.Type.Kind() == reflect.Ptr && (v == "<nil>" || v == "") {
		return nil
	}
	var tag = field.Tag.Get("gorm")
	var expected = ""
	if tag != "" {
		var enumMatch = enumRegex.FindAllStringSubmatch(tag, 1)
		if len(enumMatch) == 1 {
			var values = enumBodyRegex.FindAllStringSubmatch(enumMatch[0][1], -1)
			for _, item := range values {
				expected += "," + item[1]
				if item[1] == v {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("invalid value, expected values are: %s", strings.TrimLeft(expected, ","))
}

func afterValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	var t = value.String()
	if t == "" || strings.HasPrefix("0000-00-00", t) {
		return nil
	}
	var srcVal, err = value.Time()
	if err != nil {
		return fmt.Errorf("invalid date, date expected be in RFC3339 format")
	}
	var f, ok = stmt.Schema.FieldsByName[match[1]]
	if !ok {
		return fmt.Errorf("field %s not found", match[1])
	}
	dst, zero := f.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))
	v := reflect.ValueOf(dst)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !zero {
		if dstVal, ok := v.Interface().(time.Time); ok {
			if !srcVal.After(dstVal) {
				return fmt.Errorf("%s must be before %s", field.Name, match[1])
			}
		}
	}
	return nil
}

func beforeValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	var t = value.String()
	if t == "" || strings.HasPrefix("0000-00-00", t) {
		return nil
	}
	var srcVal, err = value.Time()
	if err != nil {
		return fmt.Errorf("invalid date, date expected be in RFC3339 format")
	}
	var f, ok = stmt.Schema.FieldsByName[match[1]]
	if !ok {
		return fmt.Errorf("field %s not found", match[1])
	}
	dst, zero := f.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))
	v := reflect.ValueOf(dst)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !zero {
		if dstVal, ok := v.Interface().(time.Time); ok {
			if !srcVal.Before(dstVal) {
				return fmt.Errorf("%s must be before %s", field.Name, match[1])
			}
		}
	}

	return nil
}

func uniqueValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	if field.StructField.Type.Kind() == reflect.Ptr && (value.String() == "<nil>" || value.String() == "") {
		return nil
	}
	if !field.Unique && value.String() == "" {
		return nil
	}
	of, zero := stmt.Schema.PrioritizedPrimaryField.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))
	var c int64
	//TODO: check for unique index
	var model = db.Table(stmt.Table).Where(field.DBName+" = ?", value.Input)
	if !zero {
		model = model.Where(stmt.Schema.PrioritizedPrimaryField.DBName+" != ?", of)
	}
	model.Count(&c)
	if c > 0 {
		return fmt.Errorf("duplicate entry")
	}
	return nil
}

func foreignKeyValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	if field.StructField.Type.Kind() == reflect.Ptr && (value.String() == "" || value.String() == "<nil>") {
		return nil
	}
	var c int64
	if foreignTable, ok := field.TagSettings["FK"]; ok {
		if foreignModel := scm.Find(foreignTable); foreignModel != nil {
			db.Where(foreignModel.PrimaryKey[0]+" = ?", value.Input).Table(foreignTable).Count(&c)
			if c == 0 {
				return fmt.Errorf("value does not match foreign key")
			}
		}
	}
	return nil
}

func textValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" {
		return nil
	}
	var re = regexp.MustCompile(`(?m)</?(\w).*\\?>`)
	if re.MatchString(v) {
		return fmt.Errorf("the text cannot contains html fields")
	}
	return nil
}

func digitValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" {
		return nil
	}

	var r = regexp.MustCompile("^[0-9]+$")
	if !r.MatchString(v) {
		return fmt.Errorf("invalid digit value: %s", v)
	}

	return nil
}

func ipValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ".")

	if len(parts) != 4 {
		return fmt.Errorf("invalid IP address: %s", v)
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return fmt.Errorf("invalid IP address: %s", v)
			}
		} else {
			return fmt.Errorf("invalid IP address: %s", v)
		}

	}
	return nil
}

func urlValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" {
		return nil
	}
	_, err := url.ParseRequestURI(v)
	if err != nil {
		return err
	}
	return nil
}

func domainValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" {
		return nil
	}
	var regex = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)
	if !regex.MatchString(v) {
		return fmt.Errorf("invalid domain: %s", v)
	}
	return nil
}

func passwordValidator(match []string, value *generic.Value) error {
	var v = value.String()
	var digit = false
	var letter = false
	var upper = false
	var symbol = false
	for _, c := range v {
		if unicode.IsDigit(c) {
			digit = true
		} else if unicode.IsLower(c) {
			letter = true
		} else if unicode.IsUpper(c) {
			upper = true
		} else if unicode.IsSymbol(c) {
			symbol = true
		}
	}
	var complexity = 0
	if digit {
		complexity += 1
	}
	if upper {
		complexity += 1
	}
	if symbol {
		complexity += 1
	}
	if letter {
		complexity += 1
	}
	switch match[1] {
	case "none", "":
		return nil
	case "easy":
		if len(v) < 6 {
			return fmt.Errorf("password must be at least 8 characters long")
		}
	case "medium":
		if len(v) < 6 {
			return fmt.Errorf("password must be at least 6 characters long")
		}
		if complexity < 3 {
			return fmt.Errorf("password is not complex enough")
		}
	case "hard":
		if len(v) < 8 {
			return fmt.Errorf("password must be at least 8 characters long")
		}
		if complexity < 5 {
			return fmt.Errorf("password is not complex enough")
		}
	}
	return nil
}

func floatValidator(match []string, value *generic.Value) error {
	var prefix = "[+-]?"
	switch match[1] {
	case "-":
		prefix = "-"
	case "+":
		prefix = "+?"
	}
	var v = value.String()
	if v == "" {
		return nil
	}
	if !regexp.MustCompile("^" + prefix + "([0-9]*[.])?[0-9]+$").MatchString(v) {
		return fmt.Errorf("invalid integer")
	}
	return nil
}

func intValidator(match []string, value *generic.Value) error {
	var prefix = "[+-]?"
	switch match[1] {
	case "-":
		prefix = "-"
	case "+":
		prefix = "+?"
	}
	var v = value.String()
	if v == "" {
		return nil
	}
	if !regexp.MustCompile("^" + prefix + "[0-9]+$").MatchString(v) {
		return fmt.Errorf("invalid integer")
	}
	return nil
}

func numericalValidator(match []string, value *generic.Value) error {
	var s = value.String()
	if s == "" || s == "<nil>" {
		return nil
	}
	var v = value.Float64()
	limit, _ := strconv.ParseFloat(match[2], 64)

	switch match[1] {
	case "<":
		if v >= limit {
			return fmt.Errorf("is bigger than %s", match[2])
		}
	case ">":
		if v <= limit {
			return fmt.Errorf("is smaller than %s", match[2])
		}
	case "<=":
		if v > limit {
			return fmt.Errorf("is bigger than or equal to %s", match[2])
		}
	case ">=":
		if v < limit {
			return fmt.Errorf("is smaller than or equal to %s", match[2])
		}
	case "==", "=":
		if v != limit {
			return fmt.Errorf("is not equal to %f", limit)
		}
	case "!=", "<>":
		if v != limit {
			return fmt.Errorf("is  equal to %f", limit)
		}
	}
	return nil
}

func lenValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	var size = len(v)
	t, _ := strconv.ParseInt(match[2], 10, 6)
	length := int(t)
	switch match[1] {
	case "<":
		if size >= length {
			return fmt.Errorf("is too long")
		}
	case ">":
		if size <= length {
			return fmt.Errorf("is too short")
		}
	case "<=":
		if size > length {
			return fmt.Errorf("is too long")
		}
	case ">=":
		if size < length {
			return fmt.Errorf("is too short")
		}
	case "==", "=":
		if size != length {
			return fmt.Errorf("is not equal to %d", size)
		}
	case "!=", "<>":
		if size == length {
			return fmt.Errorf("is  equal to %d", size)
		}
	}
	return nil
}

func nameValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if !regexp.MustCompile(`^[\p{L} .'\-]+$`).MatchString(v) {
		return fmt.Errorf("is not valid name")
	}

	return nil
}

func alphaValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	fmt.Println("alpha:", v)
	for _, r := range v {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == ' ') {
			return fmt.Errorf("is not alpha")
		}
	}
	return nil
}

func alphaNumericValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	for _, r := range v {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ') {
			return fmt.Errorf("is not alpha")
		}
	}
	return nil
}

var emailRegex = regexp.MustCompile(`(?i)^[a-z0-9_\-.]{2,}(\+\d+)?@[a-z0-9_-]{2,}\.[a-z0-9]{2,}$`)

func emailValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if strings.TrimSpace(v) == "" {
		return nil
	}
	if emailRegex.MatchString(v) {
		return nil
	}
	return fmt.Errorf("invalid email %s", v)
}

func requiredValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if strings.TrimSpace(v) == "" || v == "<nil>" {
		return fmt.Errorf("is required")
	}
	return nil
}

func regexValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if value.Input == nil || v == "" || v == "<nil>" {
		return nil
	}
	if !regexp.MustCompile(match[1]).MatchString(v) {
		return fmt.Errorf("is not valid %s", v)
	}
	return nil
}

func dateValidator(match []string, value *generic.Value) error {
	var t = value.String()
	if t == "" || strings.HasPrefix("0000-00-00", t) {
		return nil
	}
	_, err := value.Time()
	if err != nil {
		return fmt.Errorf("invalid date, date expected be in RFC3339 format")
	}
	return nil
}
