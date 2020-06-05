package ref

import (
	"fmt"
	"testing"
)

type myStruct struct {
	X string
	y int
	z int
}

func TestNew(t *testing.T) {
	x := &myStruct{
		"CDE", 9, 8,
	}

	Wrap(x).Set("X", "KKJKJJJ")

	fmt.Println(x)
}
