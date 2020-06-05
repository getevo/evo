package utils

import (
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

type String string
type RandomStringOp int

const (
	LOWER_CASE     RandomStringOp = 1
	UPPER_CASE     RandomStringOp = 2
	RANDOM_CASE    RandomStringOp = 3
	ALPHANUM       RandomStringOp = 4
	NUMBERS        RandomStringOp = 5
	ALPHANUM_SIGNS RandomStringOp = 6
)

func (s String) Trim(args ...string) String {
	if len(args) > 0 {
		for _, arg := range args {
			s = String(strings.Trim(string(s), arg))
		}
	}
	s = String(strings.TrimSpace(string(s)))
	return s
}

func (s String) TrimLeft(args ...string) String {
	if len(args) > 0 {
		for _, arg := range args {
			s = String(strings.TrimLeft(string(s), arg))
		}
	}
	s = String(strings.TrimLeft(string(s), " "))
	return s
}

func (s String) TrimRight(args ...string) String {
	if len(args) > 0 {
		for _, arg := range args {
			s = String(strings.TrimRight(string(s), arg))
		}
	}
	s = String(strings.TrimRight(string(s), " "))
	return s
}

func (s String) StartsWith(match string) bool {
	return strings.HasPrefix(string(s), match)
}

func (s String) EndsWith(match string) bool {
	return strings.HasPrefix(string(s), match)
}

func (s String) Contains(match string) bool {
	return strings.Contains(string(s), match)
}

func (s String) ContainsRegex(match string) bool {
	var re = regexp.MustCompile(match)
	return re.MatchString(string(s))
}

func (s String) Split(splitter string) []string {
	return strings.Split(string(s), splitter)
}

func (s String) SplitBy(splitter ...string) []string {
	for _, arg := range splitter {
		s = s.Replace(arg, "ø")
	}
	return strings.Split(string(s), "ø")
}

func (s String) Replace(find, replace string) String {
	return String(strings.Replace(string(s), find, replace, -1))
}

func (s String) Fields() []string {
	return strings.Fields(string(s))
}

func (s String) Title() String {
	return String(strings.Title(string(s)))
}
func (s String) TruncateWord(length int, more string) String {

	if len(s) > length {
		if length > 3 {
			length -= 3
		}
		truncated := String(s[0:length])
		if s[len(truncated)-1] != ' ' && s[len(truncated)-2] != ' ' {
			i := len(truncated)
			for i < len(s)-1 && s[i] != ' ' {
				truncated += String(s[i])
				i++
			}
		}
		truncated = truncated.TrimRight() + String(more)
		return truncated
	}
	return s
}
func (s String) Truncate(length int, more string) String {

	if len(s) > length {
		if length > 3 {
			length -= 3
		}
		s = s[0:length]
		s = s.TrimRight() + String(more)
	}
	return s
}

func (s String) isLetters() bool {
	return strings.IndexFunc(string(s), func(c rune) bool { return (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') }) == -1
}
func (s String) isDigit() bool {
	return strings.IndexFunc(string(s), func(c rune) bool { return c < '0' || c > '9' }) == -1
}

var lettersAndDigits = regexp.MustCompile("^[a-zA-Z0-9]*$")

func (s String) isLettersAndNumbers() bool {
	return lettersAndDigits.MatchString(string(s))
}

var alpha = regexp.MustCompile(`^[a-zA-Z0-9_\-]*$`)

func (s String) isAlphaNumeric() bool {
	return alpha.MatchString(string(s))
}

func (s String) String() string {
	return string(s)
}

func RandomString(length int, option RandomStringOp) string {
	var set string
	switch option {
	case LOWER_CASE:
		set = "abcdefghijklmnopqrstuvwxyz"
	case UPPER_CASE:
		set = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case RANDOM_CASE:
		set = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case ALPHANUM:
		set = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	case NUMBERS:
		set = "0123456789"
	case ALPHANUM_SIGNS:
		set = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%^&*()-_!~<>,:;"
	default:
		set = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	random := ""
	for i := 0; i < length; i++ {
		random += string(set[rand.Intn(len(set)-1)])
	}

	return random
}

func FormatInt(i int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d\n", i)
}

func FormatInt64(i int64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d\n", i)
}

func ParseSize(expr string) (int64, error) {
	expr = strings.ToLower(expr)
	num := ""
	unit := ""
	numParsed := false
	for _, char := range expr {
		if !numParsed && char >= '0' && char <= '9' {
			num += string(char)
		} else if char >= 'a' && char <= 'z' {
			unit += string(char)
		}
	}

	i, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	switch unit {
	case "", "b", "byte":
		// do nothing - already in bytes

	case "k", "kb", "kilo", "kilobyte", "kilobytes":
		return int64(uint64(i) * KB), nil

	case "m", "mb", "mega", "megabyte", "megabytes":
		return int64(uint64(i) * MB), nil

	case "g", "gb", "giga", "gigabyte", "gigabytes":
		return int64(uint64(i) * GB), nil

	case "t", "tb", "tera", "terabyte", "terabytes":
		return int64(uint64(i) * TB), nil

	case "p", "pb", "peta", "petabyte", "petabytes":
		return int64(uint64(i) * PB), nil

	case "E", "EB", "e", "eb", "eB":
		return int64(uint64(i) * EB), nil

	default:
		return 0, fmt.Errorf("unable to parse size: %s", expr)
	}
	return 0, fmt.Errorf("unable to parse size: %s", expr)
}

func ReplaceArray(input string, needles []string, replace []string) string {
	if len(needles) != len(replace) {
		return input
	}
	for k, needle := range needles {
		input = strings.Replace(input, needle, replace[k], -1)
	}
	return input
}
