package tpl

import (
	"fmt"
	"testing"
)

type User struct {
	Family string
}

func TestRender(t *testing.T) {
	var text = `Hello $name $user.Family $arr[1] $arr[2][test] $arr[2][key].Family $test`
	fmt.Println(Render(text, map[string]interface{}{
		"name": "to",
		"user": User{Family: "MH"},
		"arr":  []interface{}{"a", "b", map[string]interface{}{"test": "value", "key": User{Family: "My User"}}},
		"test": []string{"x", "y", "z"},
	}))
}
