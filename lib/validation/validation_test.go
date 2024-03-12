package validation_test

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/validation"
	"testing"
)

type Struct struct {
	Alpha          string `validation:"alpha"`
	RequiredAlpha  string `validation:"alpha,required"`
	AlphaNumeric   string `validation:"alphanumeric"`
	Email          string `validation:"email"`
	RequiredEmail  string `validation:"email,required"`
	URL            string `validation:"url"`
	Domain         string `validation:"domain"`
	IP             string `validation:"ip"`
	Regex          string `validation:"regex([a-z]{2,})"`
	LengthEq       string `validation:"len==10"`
	LengthGt       string `validation:"len>10"`
	LengthLt       string `validation:"len<10"`
	PositiveInt    string `validation:"+int"`
	NegativeInt    string `validation:"-int"`
	PositiveFloat  string `validation:"+float"`
	NegativeFloat  string `validation:"-float"`
	PositiveFloat2 string `validation:"float"`
}

func TestStruct(t *testing.T) {
	s := Struct{
		Alpha:          "a",
		RequiredAlpha:  "a",
		AlphaNumeric:   "a1",
		Email:          "reza+5@yahoo.com",
		RequiredEmail:  "reza+5@yahoo.com",
		URL:            "https://example.com",
		Domain:         "example.com",
		IP:             "127.0.0.1",
		Regex:          "ab~",
		LengthEq:       "1234567890",
		LengthGt:       "1234567890001",
		LengthLt:       "10",
		PositiveInt:    "10",
		NegativeInt:    "-10",
		PositiveFloat:  "10.5",
		NegativeFloat:  "-10.5",
		PositiveFloat2: "10.5",
	}

	errs := validation.Struct(s)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
	}
}
