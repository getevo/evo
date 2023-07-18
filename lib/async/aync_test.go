package async_test

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/async"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	var iter = async.Iterator[int]{
		MaxGoroutines: 3,
	}
	iter.ForEach([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, func(i *int) {
		fmt.Println(*i)
		time.Sleep(1 * time.Second)
	})
}
