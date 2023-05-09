package ftp

import (
	"fmt"
	"testing"
	"time"
)

func TestDriver_Init(t *testing.T) {
	var ftp = Driver{}
	fmt.Println("connection err:", ftp.Init("ftp://username:password@192.168.1.1:21/testdir"))
	fmt.Println("write testfile err:", ftp.Append("testfile", "\r\n"+time.Now().Format("15:04:03")))
	/*	fmt.Println("write testfile2 err:", ftp.Write("testfile2", "\r\n"+time.Now().Format("15:04:03")))
		fmt.Println("delete testfile2 err:", ftp.Remove("testfile2"))
		fmt.Println("delete testfile2 err:", ftp.Remove("testfile2"))*/
	//fs.Append("testfile", "\r\nline2")
	/*	fmt.Println("write testfile2 err:", ftp.Write("testfile2", "\r\n"+time.Now().Format("15:04:03")))
		fmt.Println("delete testfile2 err:", ftp.Remove("testfile2"))
		var files, err = ftp.Search("*.go")*/
	/*	fmt.Println(err)
		for _, file := range files {
			fmt.Println(file.Path())
		}*/

	//fmt.Println("testfile exists?", ftp.IsFileExists("testfile"))
	/*	fmt.Println("testfile8 exists?", ftp.IsFileExists("testfile8"))

		fmt.Println("New directory exists?", ftp.IsDirExists("New directory"))
		fmt.Println("New directory2 exists?", ftp.IsDirExists("New directory2"))*/

	/*	var stat, err = ftp.Stat("testfile")

		fmt.Println("testfile stat:", err, stat)*/
	//fmt.Println(ftp.CopyFile("testfile", "testfile22222"))
	fmt.Println(ftp.CopyDir("New directory", "./copy"))
	/*	fmt.Println(ftp.Mkdir("testdir6/6/7/8"))
		fmt.Println(ftp.MkdirAll("testdir3/2/3/4/5"))*/
}
