package network

import (
	"fmt"
	"github.com/iesreza/foundation/lib/machine"
	"testing"
)

func TestNetwork(t *testing.T) {
	res, err := GetConfig()
	fmt.Println(res, err)

}
