package settings

import "github.com/getevo/evo"

type Settings struct {
	evo.Model
	Reference string
	Title     string
	Data      string
	Default   string
	Ptr       interface{} `gorm:"-"`
}
