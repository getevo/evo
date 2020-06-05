package settings

import "github.com/iesreza/io"

type Settings struct {
	evo.Model
	Reference string
	Title     string
	Data      string
	Default   string
	Ptr       interface{} `gorm:"-"`
}
