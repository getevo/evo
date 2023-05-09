package sftp

import (
	"fmt"
	"testing"
)

func TestDriver(t *testing.T) {
	var sftp = Driver{}
	fmt.Println("connection err:", sftp.Init("sftp://ies:iesitalia2020@192.168.1.80:22//home/ies/testdir"))

	//fmt.Println(sftp.Write("testfile", "hello world"))
	//fmt.Println(sftp.Append("testfile", "123456"))

	//fmt.Println(sftp.Mkdir("./copy"))
	//fmt.Println(sftp.MkdirAll("./copy/a/b/c"))

	//fmt.Println(sftp.Remove("./testfile"))
	//fmt.Println(sftp.RemoveAll("./copy"))

	//fmt.Println("file exists?", sftp.IsFileExists("./testfile"))
	//fmt.Println(sftp.Stat("./testfile"))
	//fmt.Println(sftp.CopyFile("./testfile", "./testfile2"))
	//fmt.Println(sftp.CopyDir("./New directory", "./copy"))
	//fmt.Println(sftp.Search("*.go"))

}
