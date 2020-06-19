package T

import (
	"fmt"
	"github.com/getevo/evo/lib"
	"strconv"
)

type m string

func Must(v interface{}) m {
	return m(fmt.Sprint(v))
}

func (v m) Int() int {
	return lib.ParseSafeInt(string(v))
}

func (v m) Float() float64 {
	return lib.ParseSafeFloat(string(v))
}

func (v m) Int64() int64 {
	return lib.ParseSafeInt64(string(v))
}

func (v m) UInt() uint {
	u, _ := strconv.ParseUint(string(v), 10, 32)
	return uint(u)
}

func (v m) UInt64() uint64 {
	u, _ := strconv.ParseUint(string(v), 10, 64)
	return u
}

func (v m) Quote() string {
	return strconv.Quote(string(v))
}

func (v m) String() string {
	return fmt.Sprint(v)
}

func (v m) Bool() bool {
	return v[0] == '1' || v[0] == 't' || v[0] == 'T'
}
