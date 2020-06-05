package apiman

import (
	"fmt"
	"github.com/iesreza/io"
	"github.com/iesreza/io/menu"
	"github.com/iesreza/io/user"
)

func Register() {
	evo.Register(App{})
}

type App struct{}

var config *evo.Configuration

// Register the adminlte
func (App) Register() {
	fmt.Println("API Man Registered")
	config = evo.GetConfig()
}

// WhenReady called after setup all apps
func (App) WhenReady() {}

// Router setup routers
func (App) Router() {

}

// Permissions setup permissions of app
func (App) Permissions() []user.Permission { return []user.Permission{} }

// Menus setup menus
func (App) Menus() []menu.Menu {
	return []menu.Menu{}
}
