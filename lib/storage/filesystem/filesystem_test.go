package filesystem

import (
	"fmt"
	"testing"
)

func TestDriver_Init(t *testing.T) {
	var fs = Driver{}
	fs.Init("fs://./")
	fs.Write("testfile", "test")
	fs.Append("testfile", "\r\nline2")
	var m, err = fs.GetMetadata("Konica_Minolta_DiMAGE_Z3.jpg")
	fmt.Println(m, err)
}
