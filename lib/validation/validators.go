package validation

import (
	"context"
	"fmt"
	"github.com/getevo/evo/v2/lib/db"
	scm "github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/is"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var DBValidators = map[*regexp.Regexp]func(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error{
	regexp.MustCompile("^unique$"):          uniqueValidator,
	regexp.MustCompile("^unique:(.+)$"):     uniqueColumnsValidator,
	regexp.MustCompile("^fk$"):              foreignKeyValidator,
	regexp.MustCompile("^enum$"):            enumValidator,
	regexp.MustCompile(`^before\((\w+)\)$`): beforeValidator,
	regexp.MustCompile(`^after\((\w+)\)$`):  afterValidator,
}

func uniqueColumnsValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	var columns = strings.Split(match[1], "|")
	var model = db.Table(stmt.Table)
	var columnDbName []string
	for _, item := range stmt.Schema.Fields {
		for _, column := range columns {
			if item.DBName == column || item.Name == column {
				dst, zero := item.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))
				if zero {
					return nil
				}
				model = model.Where("`"+item.DBName+"` = ?", dst)
				columnDbName = append(columnDbName, item.DBName)
			}
		}
	}
	of, zero := stmt.Schema.PrioritizedPrimaryField.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))
	if !zero {
		model = model.Where(stmt.Schema.PrioritizedPrimaryField.DBName+" != ?", of)
	}
	var c int64
	model.Count(&c)
	if c > 0 {
		return fmt.Errorf("duplicate value for %s", strings.Join(columnDbName, ","))
	}
	return nil
}

var Validators = map[*regexp.Regexp]func(match []string, value *generic.Value) error{
	regexp.MustCompile("(?i)^text$"):                              textValidator,
	regexp.MustCompile("(?i)^name$"):                              nameValidator,
	regexp.MustCompile("(?i)^alpha$"):                             alphaValidator,
	regexp.MustCompile("(?i)^latin$"):                             latinValidator,
	regexp.MustCompile("(?i)^name$"):                              nameValidator,
	regexp.MustCompile("(?i)^digit$"):                             digitValidator,
	regexp.MustCompile("(?i)^alphanumeric$"):                      alphaNumericValidator,
	regexp.MustCompile("(?i)^required$"):                          requiredValidator,
	regexp.MustCompile("(?i)^email$"):                             emailValidator,
	regexp.MustCompile(`(?i)^regex\((.*)\)$`):                     regexValidator,
	regexp.MustCompile(`(?i)^len(>|<|<=|>=|==|!=|<>|=)(\d+)$`):    lenValidator,
	regexp.MustCompile(`(?i)^(>|<|<=|>=|==|!=|<>|=)([+\-]?\d+)$`): numericalValidator,
	regexp.MustCompile(`(?i)^([+\-]?)int$`):                       intValidator,
	regexp.MustCompile(`(?i)^([+\-]?)float$`):                     floatValidator,
	regexp.MustCompile(`(?i)^password\((.*)\)$`):                  passwordValidator,
	regexp.MustCompile(`(?i)^domain$`):                            domainValidator,
	regexp.MustCompile(`(?i)^url$`):                               urlValidator,
	regexp.MustCompile(`(?i)^ip$`):                                ipValidator,
	regexp.MustCompile(`(?i)^date$`):                              dateValidator,
	regexp.MustCompile(`(?i)^longitude`):                          longitudeValidator,
	regexp.MustCompile(`(?i)^latitude`):                           latitudeValidator,
	regexp.MustCompile(`(?i)^port$`):                              portValidator,
	regexp.MustCompile(`(?i)^json$`):                              jsonValidator,
	regexp.MustCompile(`(?i)^ISBN$`):                              isbnValidator,
	regexp.MustCompile(`(?i)^ISBN10$`):                            isbn10Validator,
	regexp.MustCompile(`(?i)^ISBN13$`):                            isbn13Validator,
	regexp.MustCompile(`(?i)^[credit[-]?card$`):                   creditCardValidator,
	regexp.MustCompile(`(?i)^uuid$`):                              uuidValidator,
	regexp.MustCompile(`(?i)^upperCase$`):                         upperCaseValidator,
	regexp.MustCompile(`(?i)^lowerCase$`):                         lowerCaseValidator,
	regexp.MustCompile(`(?i)^rgb-color$`):                         _RGBColorValidator,
	regexp.MustCompile(`(?i)^rgba-color$`):                        _RGBAColorValidator,
	regexp.MustCompile(`(?i)^hex-color$`):                         _HEXColorValidator,
	regexp.MustCompile(`(?i)^hex$`):                               _HEXValidator,
	regexp.MustCompile(`(?i)^country-alpha-2$`):                   _CountryAlpha2Validator,
	regexp.MustCompile(`(?i)^country-alpha-3`):                    _CountryAlpha3Validator,
	regexp.MustCompile(`(?i)^btc_address`):                        _BTCAddressValidator,
	regexp.MustCompile(`(?i)^eth_address`):                        _ETHAddressValidator,
	regexp.MustCompile(`(?i)^cron`):                               cronValidator,
	regexp.MustCompile(`(?i)^duration`):                           durationValidator,
	regexp.MustCompile(`(?i)^time$`):                              timestampValidator,
	regexp.MustCompile(`(?i)^(unix-timestamp|unix-ts)$`):          unixTimestampValidator,
	regexp.MustCompile(`(?i)^timezone$`):                          timezoneValidator,
	regexp.MustCompile(`(?i)^e164$`):                              e164Validator,
	regexp.MustCompile(`(?i)^safe-html`):                          safeHTMLValidator,
	regexp.MustCompile(`(?i)^no-html$`):                           noHTMLValidator,
	regexp.MustCompile(`(?i)^phone$`):                             phoneValidator,
}

func phoneValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if is.PhoneNumber(v) {
		return fmt.Errorf("value must be valid phone number")
	}
	return nil
}

func noHTMLValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if is.NoHTMLTags(v) {
		return fmt.Errorf("value must not contain any html tags")
	}
	return nil
}

func safeHTMLValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if is.SafeHTML(v) {
		return fmt.Errorf("value must not contain any possible XSS tokens")
	}
	return nil
}

func e164Validator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !e164Regex.MatchString(v) {
		return fmt.Errorf("value must be a valid E164 phone number")
	}
	return nil
}

func timezoneValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	_, err := time.LoadLocation(v)
	if err != nil {
		return fmt.Errorf("value must be a valid timezone")
	}
	return nil
}

func unixTimestampValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	_, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fmt.Errorf("value must be a valid unix timestamp")
	}

	return nil
}

func timestampValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	_, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return fmt.Errorf("value must be a valid RFC3339 timestamp")
	}
	return nil
}

func durationValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	_, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("value must be a valid duration format")
	}
	return nil
}

func cronValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.Cron(v) {
		return fmt.Errorf("value must be a valid CRON format")
	}
	return nil
}

func _BTCAddressValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	var re = regexp.MustCompile(`^(bc1|[13])[a-zA-HJ-NP-Z0-9]{25,39}$`)
	if !re.MatchString(v) {
		return fmt.Errorf("value must be a valid Bitcoin address")
	}
	return nil
}

func _ETHAddressValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	var re = regexp.MustCompile(`^(0x)[a-zA-Z0-9]{40}$`)
	if !re.MatchString(v) {
		return fmt.Errorf("value must be a valid ETH address")
	}
	return nil
}

func _CountryAlpha3Validator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.ISO3166Alpha3(v) {
		return fmt.Errorf("value must be a valid ISO3166 Alpha 3 Format")
	}
	return nil
}

func _CountryAlpha2Validator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.ISO3166Alpha3(v) {
		return fmt.Errorf("value must be a valid ISO3166 Alpha 2 Format")
	}
	return nil
}

func _HEXValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.Hexadecimal(v) {
		return fmt.Errorf("value must be a valid HEX string")
	}
	return nil
}

func _HEXColorValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.HexColor(v) {
		return fmt.Errorf("value must be HEX color")
	}
	return nil
}

func _RGBColorValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.RGBColor(v) {
		return fmt.Errorf("value must be RGB color")
	}
	return nil
}

func _RGBAColorValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.RGBAColor(v) {
		return fmt.Errorf("value must be RGBA color")
	}
	return nil
}

func lowerCaseValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.LowerCase(v) {
		return fmt.Errorf("value must be in lower case")
	}
	return nil
}

func upperCaseValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.UpperCase(v) {
		return fmt.Errorf("value must be in upper case")
	}
	return nil
}

func uuidValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.UUID(v) {
		return fmt.Errorf("value must be valid uuid")
	}
	return nil
}

func creditCardValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.CreditCard(v) {
		return fmt.Errorf("value must be credit card number")
	}
	return nil
}

func isbn13Validator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.ISBN13(v) {
		return fmt.Errorf("value must be ISBN-13 format")
	}
	return nil
}

func isbn10Validator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.ISBN10(v) {
		return fmt.Errorf("value must be ISBN-10 format")
	}
	return nil
}

func isbnValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.ISBN10(v) {
		return fmt.Errorf("value must be ISBN-10 format")
	}
	return nil
}

func jsonValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.JSON(v) {
		return fmt.Errorf("value must be valid JSON format")
	}
	return nil
}

func portValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.ISBN13(v) {
		return fmt.Errorf("value must be valid port number")
	}
	return nil
}

func latitudeValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.ISBN13(v) {
		return fmt.Errorf("value must be valid latitude")
	}
	return nil
}

func longitudeValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.ISBN13(v) {
		return fmt.Errorf("value must be valid longitude")
	}
	return nil
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
	if v == "" || v == "<nil>" {
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
	if v == "" || v == "<nil>" {
		return nil
	}

	var r = regexp.MustCompile("^[0-9]+$")
	if !r.MatchString(v) {
		return fmt.Errorf("invalid digit value")
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
		return fmt.Errorf("invalid IP address")
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return fmt.Errorf("invalid IP address")
			}
		} else {
			return fmt.Errorf("invalid IP address")
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
		return fmt.Errorf("invalid domain")
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
	var size = utf8.RuneCountInString(v)
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
	return fmt.Errorf("invalid email")
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
		return fmt.Errorf("format is not valid")
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
