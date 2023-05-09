package storage

import (
	"fmt"
	"testing"
)

func TestNewStorage(t *testing.T) {
	fmt.Println(NewStorage("filesystem", "fs://./"))
	fmt.Println(NewStorage("s3", "s3://username:password@host.tld/bucket/dir/?region=us-west-1"))

	var storage = GetStorage("s3")
	fmt.Println(storage.List("./", true))
}
