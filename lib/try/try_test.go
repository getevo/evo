package try

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/panics"
	"net"
	"testing"
)

func Test_NormalFlow(T *testing.T) {

	var x *net.IPNet
	x = nil
	This(func() {
		x.Contains(net.IP("192.168.1.0"))
	}).Catch(func(_ *panics.Recovered) {
		fmt.Println("catch received")
	})

	fmt.Println("program continues")

	This(func() {
		panic("panic!")
	}).Catch(func(err *panics.Recovered) {
		fmt.Println(err)
	})
}
