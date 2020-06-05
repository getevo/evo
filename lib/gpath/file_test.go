package gpath

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	file, err := Open("./test.file")
	if err != nil {
		log.Fatal(err)
	}
	file.Truncate()

	type A struct {
		X string
		Y string
	}
	file.WriteJson(A{"123", "456"}, true)
	fmt.Println(file.ReadAllString())
	time.Sleep(3 * time.Second)
	file.AppendString("pppp")
	time.Sleep(4 * time.Second)
	file.AppendString("ppppss")
	time.Sleep(4 * time.Second)
}
