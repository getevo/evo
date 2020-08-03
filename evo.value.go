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

	for _, item := range params {
		item = strings.TrimSpace(strings.ToLower(item))
		if item == "alpha" {
			for _, r := range s {
				if !unicode.IsLetter(r) {
					return nil
				}
			}
		} else if item == "numeric" {
			if _, err := strconv.Atoi(s); err != nil {
				return nil
			}
		} else if item == "alphanumeric" || item == "alphanum" {
			if !isAlphaNumeric(s) {
				return nil
			}
		} else if item[0:2] == ">=" {
			if cp, err := strconv.Atoi(strings.TrimSpace(item[2:])); err != nil {
				log.Error("invalid validation string %s", item)
				return nil
			} else {
				if i, err := strconv.Atoi(s); err != nil {
					return nil
				} else {
					if i < cp {
						return nil
					}
				}
			}

		} else if item[0:2] == "<=" {
			if cp, err := strconv.Atoi(strings.TrimSpace(item[2:])); err != nil {
				log.Error("invalid validation string %s", item)
				return nil
			} else {
				if i, err := strconv.Atoi(s); err != nil {
					return nil
				} else {
					if i > cp {
						return nil
					}
				}
			}
		} else if item[0] == '>' {
			if cp, err := strconv.Atoi(strings.TrimSpace(item[2:])); err != nil {
				log.Error("invalid validation string %s", item)
				return nil
			} else {
				if i, err := strconv.Atoi(s); err != nil {
					return nil
				} else {
					if i <= cp {
						return nil
					}
				}
			}
		} else if item[0] == '<' {
			if cp, err := strconv.Atoi(strings.TrimSpace(item[2:])); err != nil {
				log.Error("invalid validation string %s", item)
				return nil
			} else {
				if i, err := strconv.Atoi(s); err != nil {
					return nil
				} else {
					if i >= cp {
						return nil
					}
				}
			}
		} else if item[0:3] == "len" {
			fields := strings.Fields(item)
			if i, err := strconv.Atoi(fields[2]); err != nil {
				return nil
			} else {
				switch fields[1] {
				case ">":
					if len(s) <= i {
						return nil
					}
				case "<":
					if len(s) >= i {
						return nil
					}
				case ">=":
					if len(s) < i {
						return nil
					}
				case "<=":
					if len(s) > i {
						return nil
					}
				case "=":
					if len(s) != i {
						return nil
					}
				default:
					log.Error("invalid validation string %s", item)
					return nil

				}
			}
		} else if item[0:5] == "regex" {
			if !regexp.MustCompile(strings.TrimSpace(item[5:])).MatchString(s) {
				return nil
			}
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
