package query

import (
	"fmt"
	"github.com/iesreza/io"
	"github.com/iesreza/io/menu"
	"github.com/iesreza/io/user"
	"github.com/jinzhu/gorm"
)

var c Controller

// Register register the auth in io apps
func Register(v ...Filter) {
	if len(v) == 0 {
		evo.Register(App{})
		return
	}
	if objects.data == nil {
		objects.Init()
	}
	for _, item := range v {
		c.Register(item)
	}

}

// WhenReady called after setup all apps
func (App) WhenReady() {}

var db *gorm.DB

// App query app struct
type App struct{}

func (App) Register() {
	fmt.Println("Filter Registered")
	db = evo.GetDBO()
}

// Router setup routers
func (App) Router() {}

// Permissions setup permissions of app
func (App) Permissions() []user.Permission {
	return []user.Permission{}
}

// Menus setup menus
func (App) Menus() []menu.Menu {
	return []menu.Menu{}
}
