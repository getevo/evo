package evo

import (
	"fmt"
	"github.com/getevo/evo/lib"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type value string

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString

func Value(s string, params ...string) value {
	v := value(s)

	var _def value
	for k := len(params) - 1; k > 0; k-- {
		params[k] = strings.TrimSpace(params[k])

		if params[k] == "alpha" {
			for _, r := range s {
				if !unicode.IsLetter(r) {
					return _def
				}
			}
		} else if strings.HasPrefix(params[k], "regex") {
			if !regexp.MustCompile(strings.TrimSpace(params[k][5:])).MatchString(s) {
				return _def
			}
		} else if params[k] == "numeric" {
			if _, err := strconv.Atoi(s); err != nil {
				return _def
			}
		} else if params[k] == "alphanumeric" || params[k] == "alphanum" {
			if !isAlphaNumeric(s) {
				return _def
			}
		} else if fields := strings.Fields(params[k]); len(fields) >= 2 {

			if len(fields) == 3 && fields[0] == "len" {
				if i, err := strconv.Atoi(fields[2]); err != nil {

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
					}
				}
			} else if len(fields) == 2 {
				if i, err := strconv.Atoi(fields[1]); err != nil {
					return _def
				} else {
					if v, err := strconv.Atoi(s); err != nil {
						return _def
					} else {
						switch fields[0] {
						case ">":
							if v <= i {
								return _def
							}
						case "<":
							if v >= i {
								return _def
							}
						case ">=":
							if v < i {
								return _def
							}
						case "<=":
							if v > i {
								return _def
							}
						case "=":
							if v != i {
								return _def
							}
						}
					}
				}
			}
		} else {
			vi := value(params[k])
			_def = vi
		}
	}

	return v
}

func (v value) Int() int {
	return lib.ParseSafeInt(string(v))
}

func (v value) Float() float64 {
	return lib.ParseSafeFloat(string(v))
}

func (v value) Int64() int64 {
	return lib.ParseSafeInt64(string(v))
}

func (v value) UInt() uint {
	u, _ := strconv.ParseUint(string(v), 10, 32)
	return uint(u)
}

func (v value) UInt64() uint64 {
	u, _ := strconv.ParseUint(string(v), 10, 64)
	return u
}

func (v value) Quote() string {
	return strconv.Quote(string(v))
}

func (v value) ToString() string {
	return fmt.Sprint(v)
}

func (v value) Bool() bool {
	return v[0] == '1' || v[0] == 't' || v[0] == 'T'
}
