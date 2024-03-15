package validation

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/generic"
	"regexp"
	"strconv"
	"strings"
)

var Validators = map[*regexp.Regexp]func(match []string, value *generic.Value) error{
	regexp.MustCompile("^alpha$"):                          alphaValidator,
	regexp.MustCompile("^alphanumeric$"):                   alphaNumericValidator,
	regexp.MustCompile("^required$"):                       requiredValidator,
	regexp.MustCompile("^email$"):                          emailValidator,
	regexp.MustCompile(`^regex\((.*)\)$`):                  regexValidator,
	regexp.MustCompile(`^len(>|<|<=|>=|==|!=|<>|=)(\d+)$`): lenValidator,
	regexp.MustCompile(`^(>|<|<=|>=|==|!=|<>|=)(\d+)$`):    numericalValidator,
	regexp.MustCompile(`^([+\-]?)int$`):                    intValidator,
	regexp.MustCompile(`^([+\-]?)float$`):                  floatValidator,
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
	if s == "" {
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

func alphaValidator(match []string, value *generic.Value) error {
	var v = value.String()
	for _, r := range v {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == ' ') {
			return fmt.Errorf("is not alpha")
		}
	}
	return nil
}

func alphaNumericValidator(match []string, value *generic.Value) error {
	var v = value.String()
	for _, r := range v {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ') {
			return fmt.Errorf("is not alpha")
		}
	}
	return nil
}

var emailRegex = regexp.MustCompile(`(?i)^[a-z0-9_-]{2,}(\+\d+)?@[a-z0-9_-]{2,}\.[a-z0-9]{2,}$`)

func emailValidator(match []string, value *generic.Value) error {
	var v = value.String()
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
	if strings.TrimSpace(v) == "" {
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
