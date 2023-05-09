package s3

import (
	"fmt"
	"testing"
)

var config = "s3://username:password@host.tld/bucket/dir/?region=us-west-1"

func TestDriver_Write(t *testing.T) {
	var s3 = Driver{}
	fmt.Println(s3.Init(config))

	//fmt.Println(s3.Write("./././myfile.txt", "hello world"))

	//fmt.Println(s3.Write("mydir/myfile.txt", "hello world 2"))

	/*	fmt.Println("is dir mydir?", s3.IsDirExists("mydir"))
		fmt.Println("is dir mydir2?", s3.IsDirExists("mydir2"))

		var ts = time.Now().Format("2006-01-02-15-04-03")
		fmt.Println("create dir?"+"mydir/"+ts, s3.Mkdir("mydir/"+ts))
		fmt.Println("exists dir?"+"mydir/"+ts, s3.IsDirExists("mydir/"+ts))

		fmt.Println("exists mydir/myfile.txt?", s3.IsFileExists("mydir/myfile.txt"))

		fmt.Println("IS DIR EXISTS?"+"mydir/myfile.txt"+ts, s3.IsDirExists("mydir/myfile.txt"))

		fmt.Println("path is dir?"+"mydir/myfile.txt"+ts, s3.IsDir("mydir/myfile.txt"))
		fmt.Println("path is dir?"+"mydir/"+ts, s3.IsDir("mydir/"))

		var s, _ = s3.Stat("mydir/myfile.txt")
		fmt.Println(s.Name(), s.Size(), s.ModTime())
		fmt.Println(s3.SetMetadata("mydir/myfile.txt", map[string]string{
			"hello": "world",
		}))

		fmt.Println(s3.CopyFile("./mydir/myfile.txt", "mydir/myfile.copy.txt"))
		fmt.Println(s3.GetMetadata("mydir/myfile.txt"))*/
	/*	fmt.Println("create mydir/file_to_delete.txt", s3.Touch("mydir/file_to_delete.txt"))
		fmt.Println(s3.CopyDir("./mydir", "./copy"))

		fmt.Println(s3.Remove("mydir/file_to_delete.txt"))
		fmt.Println(s3.RemoveAll("copy"))*/

	var files, err = s3.List("./", true)
	fmt.Println(err)
	for _, file := range files {
		fmt.Println(file.Path(), "is dir?", file.IsDir())
	}
}
