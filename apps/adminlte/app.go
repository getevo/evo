package adminlte

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/apps/settings"
	"github.com/getevo/evo/menu"
	"github.com/getevo/evo/viewfn"
)

// Register register the adminlte in io apps
func Register() {
	evo.Register(App{})
}

//Path to adminlte app
var Path string

// App adminlte app struct
type App struct{}

var setting = &Settings{
	NavbarColor:     "green",
	NavbarVariation: "dark",
	SidebarColor:    "green",
}
var pages *jet.Set
var elements *jet.Set
var config *evo.Configuration

// Register the adminlte
func (App) Register() {
	fmt.Println("AdminLTE Registered")
	Path = evo.GuessAsset(App{})
	pages = evo.RegisterView("template", Path+"/pages")
	elements = evo.RegisterView("html", Path+"/html")
	config = evo.GetConfig()
	settings.Register("AdminLTE Template", setting)

}

// WhenReady called after setup all apps
func (App) WhenReady() {
	pages.AddGlobal("title", config.App.Name)
	pages.AddGlobal("nav", evo.AppMenus)
	pages.AddGlobal("settings", setting)
	viewfn.Bind(pages, "thumb")
}

// Router setup routers
func (App) Router() {
	evo.Static("/assets", Path+"/assets")
	evo.Static("/plugins", Path+"/plugins")

	evo.Get("", func(r *evo.Request) {

		r.Var("heading", "Test")
		r.View(nil, "template.default")
	})

	evo.Get("/test", func(r *evo.Request) {

		r.Var("heading", "Test1")
		r.View(nil, "template.default")
	})

	evo.Get("/test2", func(r *evo.Request) {

		r.Var("heading", "Test2")
		r.View(nil, "template.default")
	})
}

// Permissions setup permissions of app
func (App) Permissions() []evo.Permission { return []evo.Permission{} }

// Menus setup menus
func (App) Menus() []menu.Menu {
	return []menu.Menu{}
}

func (App) Pack() {
	evo.Pack(Path)
}
