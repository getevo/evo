package main

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/storage"
	"testing"
)

func TestStorage(t *testing.T) {
	fmt.Println(storage.NewStorage("filesystem", "fs://./"))
	fmt.Println(storage.NewStorage("s3", "s3://admin:iesitalia2020@s3-sslazio-staging.hiwaymedia.it/sslazio/reza/?region=us-west-1"))
	fmt.Println(storage.NewStorage("ftp", "ftp://ies:iesitalia2020@192.168.1.80:21/testdir"))
	fmt.Println(storage.NewStorage("sftp", "s3://admin:iesitalia2020@s3-sslazio-staging.hiwaymedia.it/sslazio/reza/?region=us-west-1"))

}
