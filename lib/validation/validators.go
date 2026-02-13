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
	"net"
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
	regexp.MustCompile("^unique$"):                uniqueValidator,
	regexp.MustCompile("^unique:(.+)$"):           uniqueColumnsValidator,
	regexp.MustCompile("^fk$"):                    foreignKeyValidator,
	regexp.MustCompile("^enum$"):                  enumValidator,
	regexp.MustCompile(`^before\((\w+)\)$`):       beforeValidator,
	regexp.MustCompile(`^after\((\w+)\)$`):        afterValidator,
	regexp.MustCompile(`^confirmed$`):             confirmedValidator,
	regexp.MustCompile(`^same\((\w+)\)$`):         sameValidator,
	regexp.MustCompile(`^different\((\w+)\)$`):    differentValidator,
	regexp.MustCompile(`^gt[-_]?field\((\w+)\)$`): gtFieldValidator,
	regexp.MustCompile(`^gte[-_]?field\((\w+)\)$`): gteFieldValidator,
	regexp.MustCompile(`^lt[-_]?field\((\w+)\)$`):  ltFieldValidator,
	regexp.MustCompile(`^lte[-_]?field\((\w+)\)$`): lteFieldValidator,
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
	regexp.MustCompile("(?i)^slug$"):                              slugValidator,
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
	regexp.MustCompile(`(?i)^ip(v?4)?$`):                          ip4Validator,
	regexp.MustCompile(`(?i)^ipv?6$`):                             ip6Validator,
	regexp.MustCompile(`(?i)^cidr$`):                              cidrValidator,
	regexp.MustCompile(`(?i)^mac$`):                               macValidator,
	regexp.MustCompile(`(?i)^date$`):                              dateValidator,
	regexp.MustCompile(`(?i)^longitude`):                          longitudeValidator,
	regexp.MustCompile(`(?i)^latitude`):                           latitudeValidator,
	regexp.MustCompile(`(?i)^port$`):                              portValidator,
	regexp.MustCompile(`(?i)^json$`):                              jsonValidator,
	regexp.MustCompile(`(?i)^ISBN$`):                              isbnValidator,
	regexp.MustCompile(`(?i)^ISBN10$`):                            isbn10Validator,
	regexp.MustCompile(`(?i)^ISBN13$`):                            isbn13Validator,
	regexp.MustCompile(`(?i)^credit[-_]?card$`):                   creditCardValidator,
	regexp.MustCompile(`(?i)^uuid$`):                              uuidValidator,
	regexp.MustCompile(`(?i)^upperCase$`):                         upperCaseValidator,
	regexp.MustCompile(`(?i)^lowerCase$`):                         lowerCaseValidator,
	regexp.MustCompile(`(?i)^rgb[-_]?color$`):                     _RGBColorValidator,
	regexp.MustCompile(`(?i)^rgba[-_]?color$`):                    _RGBAColorValidator,
	regexp.MustCompile(`(?i)^hex[-_]?color$`):                     _HEXColorValidator,
	regexp.MustCompile(`(?i)^hex$`):                               _HEXValidator,
	regexp.MustCompile(`(?i)^country[-_]?alpha[-_]?2$`):           _CountryAlpha2Validator,
	regexp.MustCompile(`(?i)^country[-_]?alpha[-_]?3`):            _CountryAlpha3Validator,
	regexp.MustCompile(`(?i)^btc[-_]?address`):                    _BTCAddressValidator,
	regexp.MustCompile(`(?i)^eth[-_]?address`):                    _ETHAddressValidator,
	regexp.MustCompile(`(?i)^cron`):                               cronValidator,
	regexp.MustCompile(`(?i)^duration`):                           durationValidator,
	regexp.MustCompile(`(?i)^time$`):                              timestampValidator,
	regexp.MustCompile(`(?i)^(unix[-_]?timestamp|unix[-_]?ts)$`):  unixTimestampValidator,
	regexp.MustCompile(`(?i)^timezone$`):                          timezoneValidator,
	regexp.MustCompile(`(?i)^e164$`):                              e164Validator,
	regexp.MustCompile(`(?i)^safe[-_]?html`):                      safeHTMLValidator,
	regexp.MustCompile(`(?i)^no[-_]?html$`):                       noHTMLValidator,
	regexp.MustCompile(`(?i)^phone$`):                             phoneValidator,
	regexp.MustCompile(`(?i)^in\((.+)\)$`):                       inValidator,
	regexp.MustCompile(`(?i)^not[-_]?in\((.+)\)$`):               notInValidator,
	regexp.MustCompile(`(?i)^contains\((.+)\)$`):                 containsValidator,
	regexp.MustCompile(`(?i)^not[-_]?contains\((.+)\)$`):         notContainsValidator,
	regexp.MustCompile(`(?i)^starts[-_]?with\((.+)\)$`):          startsWithValidator,
	regexp.MustCompile(`(?i)^ends[-_]?with\((.+)\)$`):            endsWithValidator,
	regexp.MustCompile(`(?i)^min[-_]?items\((\d+)\)$`):           minItemsValidator,
	regexp.MustCompile(`(?i)^max[-_]?items\((\d+)\)$`):           maxItemsValidator,
	regexp.MustCompile(`(?i)^unique[-_]?items$`):                 uniqueItemsValidator,
	regexp.MustCompile(`(?i)^ascii$`):                            asciiValidator,
	regexp.MustCompile(`(?i)^printable$`):                        printableValidator,
	regexp.MustCompile(`(?i)^before[-_]?now$`):                   beforeNowValidator,
	regexp.MustCompile(`(?i)^after[-_]?now$`):                    afterNowValidator,
	regexp.MustCompile(`(?i)^date[-_]?format\((.+)\)$`):          dateFormatValidator,
	regexp.MustCompile(`(?i)^iban$`):                             ibanValidator,
}

func slugValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	var re = regexp.MustCompile(`(?m)^[a-z0-9_-]{1,200}$`)
	if !re.MatchString(v) {
		return fmt.Errorf("slug can contain only lowercase letters, numbers, hyphens, underscores, and must be between 1 and 200 characters long")
	}
	return nil
}

func macValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	macRegex := regexp.MustCompile(`(?i)^([0-9A-F]{2}:){5}[0-9A-F]{2}$`)
	if !macRegex.MatchString(v) {
		return fmt.Errorf("value must be valid MAC address")
	}
	return nil
}

func cidrValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	_, _, err := net.ParseCIDR(v)
	if err != nil {
		return fmt.Errorf("value must be valid CIDR notation")
	}
	return nil
}

func ip6Validator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	ip := net.ParseIP(v)
	if ip == nil {
		return fmt.Errorf("value must be valid IPv6 address")
	}
	// net.IP is either 4 or 16 bytes. 16 bytes means it's IPv6.
	if ip.To16() == nil && ip.To4() != nil {
		return fmt.Errorf("value must be valid IPv6 address")
	}
	return nil
}

func ip4Validator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	ip := net.ParseIP(v)
	if ip == nil {
		return fmt.Errorf("value must be valid IPv4 address")
	}
	// net.IP is either 4 or 16 bytes. 16 bytes means it's IPv6.
	if ip.To16() != nil && ip.To4() == nil {
		return fmt.Errorf("value must be valid IPv4 address")
	}
	return nil
}

func phoneValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !is.PhoneNumber(v) {
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
	if !is.ISO3166Alpha2(v) {
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
	port, err := strconv.Atoi(v)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("value must be valid port number (1-65535)")
	}
	return nil
}

func latitudeValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	lat, err := strconv.ParseFloat(v, 64)
	if err != nil || lat < -90 || lat > 90 {
		return fmt.Errorf("value must be valid latitude (-90 to 90)")
	}
	return nil
}

func longitudeValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	lng, err := strconv.ParseFloat(v, 64)
	if err != nil || lng < -180 || lng > 180 {
		return fmt.Errorf("value must be valid longitude (-180 to 180)")
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
		} else if unicode.IsSymbol(c) || unicode.IsPunct(c) {
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
			return fmt.Errorf("password must be at least 6 characters long")
		}
	case "medium":
		if len(v) < 8 {
			return fmt.Errorf("password must be at least 8 characters long")
		}
		if complexity < 3 {
			return fmt.Errorf("password must contain at least 3 of: uppercase, lowercase, digits, symbols")
		}
	case "hard":
		if len(v) < 12 {
			return fmt.Errorf("password must be at least 12 characters long")
		}
		if complexity < 4 {
			return fmt.Errorf("password must contain uppercase, lowercase, digits, and symbols")
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
		if v == limit {
			return fmt.Errorf("is equal to %f", limit)
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

var emailRegex = regexp.MustCompile(`(?i)^[a-z0-9_\-.]{2,}(\+\d+)?@([a-z0-9_-]{2,}\.)+[a-z0-9]{2,}$`)

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

// New String Validators

func inValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	values := strings.Split(match[1], ",")
	for _, val := range values {
		if strings.TrimSpace(val) == v {
			return nil
		}
	}
	return fmt.Errorf("value must be one of: %s", match[1])
}

func notInValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	values := strings.Split(match[1], ",")
	for _, val := range values {
		if strings.TrimSpace(val) == v {
			return fmt.Errorf("value must not be one of: %s", match[1])
		}
	}
	return nil
}

func containsValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !strings.Contains(v, match[1]) {
		return fmt.Errorf("value must contain '%s'", match[1])
	}
	return nil
}

func notContainsValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if strings.Contains(v, match[1]) {
		return fmt.Errorf("value must not contain '%s'", match[1])
	}
	return nil
}

func startsWithValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !strings.HasPrefix(v, match[1]) {
		return fmt.Errorf("value must start with '%s'", match[1])
	}
	return nil
}

func endsWithValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	if !strings.HasSuffix(v, match[1]) {
		return fmt.Errorf("value must end with '%s'", match[1])
	}
	return nil
}

func asciiValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	for _, r := range v {
		if r > unicode.MaxASCII {
			return fmt.Errorf("value must contain only ASCII characters")
		}
	}
	return nil
}

func printableValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	for _, r := range v {
		if !unicode.IsPrint(r) {
			return fmt.Errorf("value must contain only printable characters")
		}
	}
	return nil
}

// Array/Slice Validators

func minItemsValidator(match []string, value *generic.Value) error {
	val := reflect.ValueOf(value.Input)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil
	}
	min, _ := strconv.Atoi(match[1])
	if val.Len() < min {
		return fmt.Errorf("must have at least %d items", min)
	}
	return nil
}

func maxItemsValidator(match []string, value *generic.Value) error {
	val := reflect.ValueOf(value.Input)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil
	}
	max, _ := strconv.Atoi(match[1])
	if val.Len() > max {
		return fmt.Errorf("must have at most %d items", max)
	}
	return nil
}

func uniqueItemsValidator(match []string, value *generic.Value) error {
	val := reflect.ValueOf(value.Input)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil
	}
	seen := make(map[interface{}]bool)
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		if seen[item] {
			return fmt.Errorf("all items must be unique")
		}
		seen[item] = true
	}
	return nil
}

// Date/Time Validators

func beforeNowValidator(match []string, value *generic.Value) error {
	var t = value.String()
	if t == "" || strings.HasPrefix("0000-00-00", t) {
		return nil
	}
	dateVal, err := value.Time()
	if err != nil {
		return fmt.Errorf("invalid date format")
	}
	if !dateVal.Before(time.Now()) {
		return fmt.Errorf("date must be in the past")
	}
	return nil
}

func afterNowValidator(match []string, value *generic.Value) error {
	var t = value.String()
	if t == "" || strings.HasPrefix("0000-00-00", t) {
		return nil
	}
	dateVal, err := value.Time()
	if err != nil {
		return fmt.Errorf("invalid date format")
	}
	if !dateVal.After(time.Now()) {
		return fmt.Errorf("date must be in the future")
	}
	return nil
}

func dateFormatValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	_, err := time.Parse(match[1], v)
	if err != nil {
		return fmt.Errorf("date must be in format: %s", match[1])
	}
	return nil
}

// Financial Validators

func ibanValidator(match []string, value *generic.Value) error {
	var v = value.String()
	if v == "" || v == "<nil>" {
		return nil
	}
	// Basic IBAN validation (2 letter country code + 2 check digits + up to 30 alphanumeric)
	// IBAN must be uppercase
	re := regexp.MustCompile(`^[A-Z]{2}\d{2}[A-Z0-9]{1,30}$`)
	if !re.MatchString(v) {
		return fmt.Errorf("value must be a valid IBAN")
	}
	return nil
}

// Cross-Field Validators (DB validators with access to other fields)

func confirmedValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	// Looks for field_confirmation field
	confirmFieldName := field.Name + "Confirmation"
	confirmField, ok := stmt.Schema.FieldsByName[confirmFieldName]
	if !ok {
		return fmt.Errorf("confirmation field '%s' not found", confirmFieldName)
	}

	originalValue, _ := field.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))
	confirmValue, _ := confirmField.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))

	if originalValue != confirmValue {
		return fmt.Errorf("confirmation does not match")
	}
	return nil
}

func sameValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	otherField, ok := stmt.Schema.FieldsByName[match[1]]
	if !ok {
		return fmt.Errorf("field %s not found", match[1])
	}

	thisValue, _ := field.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))
	otherValue, _ := otherField.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))

	if thisValue != otherValue {
		return fmt.Errorf("must be same as %s", match[1])
	}
	return nil
}

func differentValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	otherField, ok := stmt.Schema.FieldsByName[match[1]]
	if !ok {
		return fmt.Errorf("field %s not found", match[1])
	}

	thisValue, _ := field.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))
	otherValue, _ := otherField.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))

	if thisValue == otherValue {
		return fmt.Errorf("must be different from %s", match[1])
	}
	return nil
}

func gtFieldValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	return compareFields(value, stmt, field, match[1], ">")
}

func gteFieldValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	return compareFields(value, stmt, field, match[1], ">=")
}

func ltFieldValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	return compareFields(value, stmt, field, match[1], "<")
}

func lteFieldValidator(match []string, value *generic.Value, stmt *gorm.Statement, field *schema.Field) error {
	return compareFields(value, stmt, field, match[1], "<=")
}

func compareFields(value *generic.Value, stmt *gorm.Statement, field *schema.Field, otherFieldName, operator string) error {
	otherField, ok := stmt.Schema.FieldsByName[otherFieldName]
	if !ok {
		return fmt.Errorf("field %s not found", otherFieldName)
	}

	thisVal := value.Float64()
	otherValue, _ := otherField.ValueOf(context.Background(), reflect.ValueOf(stmt.Model))
	otherVal := reflect.ValueOf(otherValue).Float()

	switch operator {
	case ">":
		if thisVal <= otherVal {
			return fmt.Errorf("must be greater than %s", otherFieldName)
		}
	case ">=":
		if thisVal < otherVal {
			return fmt.Errorf("must be greater than or equal to %s", otherFieldName)
		}
	case "<":
		if thisVal >= otherVal {
			return fmt.Errorf("must be less than %s", otherFieldName)
		}
	case "<=":
		if thisVal > otherVal {
			return fmt.Errorf("must be less than or equal to %s", otherFieldName)
		}
	}
	return nil
}
