package sanitize

import (
	"fmt"
	"testing"
)

type B struct {
	RN interface{}
}

type M map[string]interface{}

func TestGeneric(t *testing.T) {
	r := "' test"
	s := M{}
	s["A"] = r
	s["B"] = &r
	s["c"] = B{r}
	s["d"] = B{&r}

	Generic(&s)
	fmt.Println(s["c"].(B).RN)
}
