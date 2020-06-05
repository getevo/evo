package network

import (
	"fmt"
	"testing"
)

func TestNetwork(t *testing.T) {
	res, err := GetConfig()
	fmt.Println(res, err)

}
