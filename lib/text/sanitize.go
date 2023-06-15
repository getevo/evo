package text

import "regexp"

type Allowed *regexp.Regexp

var (
	Alpha   Allowed = regexp.MustCompile(`([a-zA-Z\s])`)
	Numeric Allowed = regexp.MustCompile(`([0-9])`)
	Decimal Allowed = regexp.MustCompile(`^[+-]?([0-9]*[.])?[0-9]+$`)
)

func Sanitize() {

}
