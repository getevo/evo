package telegram

import (
	"fmt"
	"github.com/CloudyKit/jet"
	"github.com/gofiber/fiber"
	"github.com/iesreza/io"
	"github.com/iesreza/io/apps/settings"
	"github.com/iesreza/io/lib/gpath"
	"github.com/iesreza/io/menu"
	"github.com/iesreza/io/user"
	"github.com/iesreza/io/viewfn"
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
	Path = gpath.Parent(gpath.WorkingDir()) + "/apps/adminlte"
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

	evo.Get("", func(ctx *fiber.Ctx) {
		r := evo.Upgrade(ctx)
		r.Var("heading", "Test")
		r.View(nil, "template.default")
	})

	evo.Get("/test", func(ctx *fiber.Ctx) {
		r := evo.Upgrade(ctx)
		r.Var("heading", "Test1")
		r.View(nil, "template.default")
	})

	evo.Get("/test2", func(ctx *fiber.Ctx) {
		r := evo.Upgrade(ctx)
		r.Var("heading", "Test2")
		r.View(nil, "template.default")
	})
}

// Permissions setup permissions of app
func (App) Permissions() []user.Permission { return []user.Permission{} }

// Menus setup menus
func (App) Menus() []menu.Menu {
	return []menu.Menu{}
}
