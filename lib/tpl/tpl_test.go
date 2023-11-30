package tpl

import (
	"fmt"
	"testing"
)

type User struct {
	Name   string
	Family string
}

func TestRender(t *testing.T) {
	var text = `Hello $title $user.Name $user.Family you have $sender[0] email From $sender[2][from]($sender[2][user].Name $sender[2][user].Family) at $date[0]:$date[1]:$date[2]`
	fmt.Println(Render(text, map[string]any{
		"title":  "Mrs",
		"user":   User{Name: "Maria", Family: "Rossy"},
		"sender": []any{1, "empty!", map[string]any{"from": "example@example.com", "user": User{Name: "Marco", Family: "Pollo"}}},
		"date":   []int{10, 15, 20},
	}))
}
