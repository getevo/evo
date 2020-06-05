package test

import (
	"github.com/getevo/evo"
)

type MyModel struct {
	evo.Model
	Name     string
	Username string
	Group    int
	Type     int
	Alias    string
}

type MyGroup struct {
	evo.Model
	Name string
}
