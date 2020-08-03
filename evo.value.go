package evo

import (
	"fmt"
	"github.com/getevo/evo/lib"
	"github.com/getevo/evo/lib/log"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type value string

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString

func Value(s string, params ...string) *value {
	v := value(s)
	var _def *value
	for k := len(params) - 1; k > 0; k-- {
		item := params[k]
		item = strings.TrimSpace(strings.ToLower(item))

		if item == "alpha" {
			for _, r := range s {
				if !unicode.IsLetter(r) {
					return _def
				}
			}
		} else if item == "numeric" {
			if _, err := strconv.Atoi(s); err != nil {
				return _def
			}
		} else if item == "alphanumeric" || item == "alphanum" {
			if !isAlphaNumeric(s) {
				return _def
			}
		} else if len(item) > 2 && item[0:2] == ">=" {
			if cp, err := strconv.Atoi(strings.TrimSpace(item[2:])); err != nil {
				log.Error("invalid validation string %s", item)
				return _def
			} else {
				if i, err := strconv.Atoi(s); err != nil {
					return _def
				} else {
					if i < cp {
						return _def
					}
				}
			}

		} else if len(item) > 2 && item[0:2] == "<=" {
			if cp, err := strconv.Atoi(strings.TrimSpace(item[2:])); err != nil {
				log.Error("invalid validation string %s", item)
				return _def
			} else {
				if i, err := strconv.Atoi(s); err != nil {
					return _def
				} else {
					if i > cp {
						return _def
					}
				}
			}
		} else if len(item) > 1 && item[0] == '>' {
			if cp, err := strconv.Atoi(strings.TrimSpace(item[2:])); err != nil {
				log.Error("invalid validation string %s", item)
				return _def
			} else {
				if i, err := strconv.Atoi(s); err != nil {
					return _def
				} else {
					if i <= cp {
						return _def
					}
				}
			}
		} else if len(item) > 1 && item[0] == '<' {

			if cp, err := strconv.Atoi(strings.TrimSpace(item[2:])); err != nil {
				log.Error("invalid validation string %s", item)
				return _def
			} else {
				if i, err := strconv.Atoi(s); err != nil {
					return _def
				} else {
					if i >= cp {
						return _def
					}
				}
			}
		} else if len(item) > 3 && item[0:3] == "len" {
			fields := strings.Fields(item)
			if i, err := strconv.Atoi(fields[2]); len(fields) != 3 || err != nil {
				return _def
			} else {
				switch fields[1] {
				case ">":
					if len(s) <= i {
						return _def
					}
				case "<":
					if len(s) >= i {
						return _def
					}
				case ">=":
					if len(s) < i {
						return _def
					}
				case "<=":
					if len(s) > i {
						return _def
					}
				case "=":
					if len(s) != i {
						return _def
					}
				default:
					log.Error("invalid validation string %s", item)
					return _def

				}
			}
		} else if len(item) > 5 && item[0:5] == "regex" {
			if !regexp.MustCompile(strings.TrimSpace(item[5:])).MatchString(s) {
				return _def
			}
		} else if k == len(params) {
			v := value(item)
			_def = &v
		}

	}

	return &v
}

func (v *value) Int() int {
	return lib.ParseSafeInt(string(*v))
}

func (v *value) Float() float64 {
	return lib.ParseSafeFloat(string(*v))
}

func (v *value) Int64() int64 {
	return lib.ParseSafeInt64(string(*v))
}

func (v *value) UInt() uint {
	u, _ := strconv.ParseUint(string(*v), 10, 32)
	return uint(u)
}

func (v *value) UInt64() uint64 {
	u, _ := strconv.ParseUint(string(*v), 10, 64)
	return u
}

func (v *value) Quote() string {
	return strconv.Quote(string(*v))
}

func (v *value) String() string {
	return fmt.Sprint(v)
}

func (v *value) Bool() bool {
	return (*v)[0] == '1' || (*v)[0] == 't' || (*v)[0] == 'T'
}
