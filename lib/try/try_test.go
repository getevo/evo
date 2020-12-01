package try

import (
	"fmt"
	"net"
	"testing"
)

func Test_NormalFlow(T *testing.T) {

	var x *net.IPNet
	x = nil
	This(func() {
		x.Contains(net.IP("192.168.1.0"))
	}).Catch(func(_ Error) {
		fmt.Println("catch received")
	})

	fmt.Println("program continues")

}
