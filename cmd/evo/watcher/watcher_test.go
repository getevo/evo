package watcher

import (
	"fmt"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	NewWatcher("./", func() {
		fmt.Println("change detected")
	})
	for {
		time.Sleep(1 * time.Second)
	}
}
