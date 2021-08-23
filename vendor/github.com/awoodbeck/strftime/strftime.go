// Package strftime provides a C99-compatible strftime formatter for use with
// Go time.Time instances.
package strftime

import (
	"bytes"
	"fmt"
	"strings"
	"text/scanner"
	"time"
	"unicode"
)

// Format accepts a Time pointer and a C99-compatible strftime format string
// and returns the formatted result.  If the Time pointer is nil, the current
// time is used.
//
// See http://en.cppreference.com/w/c/chrono/strftime for available conversion
// specifiers.  Specific locale specifiers are not yet supported.
func Format(t *time.Time, f string) string {
	var (
		buf bytes.Buffer
		s   scanner.Scanner
	)

	if t == nil {
		now := time.Now()
		t = &now
	}

	s.Init(strings.NewReader(f))
	s.IsIdentRune = func(ch rune, i int) bool {
		return (ch == '%' && i <= 1) || (unicode.IsLetter(ch) && i == 1)
	}

	// Honor all white space characters.
	s.Whitespace = 0

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		txt := s.TokenText()
		if len(txt) < 2 || !strings.HasPrefix(txt, "%") {
			buf.WriteString(txt)

			continue
		}

		buf.WriteString(formats.Apply(*t, txt[1:]))
	}

	return buf.String()
}

// formatMap allows for quick lookup of supported formats where the key
// is the C99-compatible conversion specifier and the value is a function
// that accepts a Time object and returns the formatted datetime string.
type formatMap map[string]func(time.Time) string

// Apply the C99-compatible conversion specifier to the Time object and
// return the formatted datetime string.  If no such specifier exists,
// the original conversion specifier prepended with the delimiter will
// be returned.
func (f formatMap) Apply(t time.Time, txt string) string {
	fc, ok := f[txt]
	if !ok {
		return fmt.Sprintf("%%%s", txt)
	}

	return fc(t)
}

var formats = formatMap{
	"a": func(t time.Time) string { return t.Format("Mon") },
	"A": func(t time.Time) string { return t.Format("Monday") },
	"b": func(t time.Time) string { return t.Format("Jan") },
	"B": func(t time.Time) string { return t.Format("January") },
	"c": func(t time.Time) string { return t.Format(time.ANSIC) },
	"C": func(t time.Time) string { return t.Format("2006")[:2] },
	"d": func(t time.Time) string { return t.Format("02") },
	"D": func(t time.Time) string { return t.Format("01/02/06") },
	"e": func(t time.Time) string { return t.Format("_2") },
	"F": func(t time.Time) string { return t.Format("2006-01-02") },
	"g": func(t time.Time) string {
		y, _ := t.ISOWeek()
		return fmt.Sprintf("%d", y)[2:]
	},
	"G": func(t time.Time) string {
		y, _ := t.ISOWeek()
		return fmt.Sprintf("%d", y)
	},
	"h": func(t time.Time) string { return t.Format("Jan") },
	"H": func(t time.Time) string { return t.Format("15") },
	"I": func(t time.Time) string { return t.Format("03") },
	"j": func(t time.Time) string { return fmt.Sprintf("%03d", t.YearDay()) },
	"k": func(t time.Time) string { return fmt.Sprintf("%2d", t.Hour()) },
	"l": func(t time.Time) string { return fmt.Sprintf("%2s", t.Format("3")) },
	"m": func(t time.Time) string { return t.Format("01") },
	"M": func(t time.Time) string { return t.Format("04") },
	"n": func(t time.Time) string { return "\n" },
	"p": func(t time.Time) string { return t.Format("PM") },
	"P": func(t time.Time) string { return t.Format("pm") },
	"r": func(t time.Time) string { return t.Format("03:04:05 PM") },
	"R": func(t time.Time) string { return t.Format("15:04") },
	"s": func(t time.Time) string { return fmt.Sprintf("%d", t.Unix()) },
	"S": func(t time.Time) string { return t.Format("05") },
	"t": func(t time.Time) string { return "\t" },
	"T": func(t time.Time) string { return t.Format("15:04:05") },
	"u": func(t time.Time) string {
		d := t.Weekday()
		if d == 0 {
			d = 7
		}
		return fmt.Sprintf("%d", d)
	},
	// "U": func(t time.Time) string {
	// TODO
	// },
	"V": func(t time.Time) string {
		_, w := t.ISOWeek()
		return fmt.Sprintf("%02d", w)
	},
	"w": func(t time.Time) string {
		return fmt.Sprintf("%d", t.Weekday())
	},
	// "W": func(t time.Time) string {
	// TODO
	// },
	"x": func(t time.Time) string { return t.Format("01/02/2006") },
	"X": func(t time.Time) string { return t.Format("15:04:05") },
	"y": func(t time.Time) string { return t.Format("06") },
	"Y": func(t time.Time) string { return t.Format("2006") },
	"z": func(t time.Time) string { return t.Format("-0700") },
	"Z": func(t time.Time) string { return t.Format("MST") },
	"%": func(t time.Time) string { return "%" },
}
