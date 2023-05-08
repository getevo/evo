package storage

import (
	"fmt"
	"testing"
)

func TestNewStorage(t *testing.T) {
	fmt.Println(NewStorage("filesystem", "fs://./"))
	fmt.Println(NewStorage("s3", "s3://admin:iesitalia2020@eu-west-1@s3-sslazio-staging.hiwaymedia.it/sslazio/VMFS1/FILES/public/upload/?region=us-west-1"))
}
