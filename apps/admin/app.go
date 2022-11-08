package admin

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/apps/auth"
	"github.com/getevo/evo/apps/settings"
	"github.com/getevo/evo/lib/fontawesome"
	"github.com/getevo/evo/menu"
	"github.com/iesreza/jet/v8"
)

// Register register the admin in io apps
func Register() {
	fmt.Println("Dashboard Registered")
	evo.Register(App{})
}

var db = evo.GetDBO()

//Path to admin app
var Path string

// App admin app struct
type App struct{}

func (App) Pack() {
	evo.Pack(Path)
}

var views *jet.Set
var setting Settings

// Register register the app
func (App) Register() {
	//Require auth
	setting.SessionAge = fmt.Sprint(evo.GetConfig().JWT.Age.Seconds())
	Path = evo.GuessAsset(App{})
	auth.Register()
	views = evo.RegisterView("admin", Path+"/views")
	settings.Register("Admin Panel", &setting)
}

// Router setup routers
func (App) Router() {
	evo.Get("/admin/login", func(r *evo.Request) {
		r.View(nil, "admin.login", "template.login")
	})

	evo.Get("/admin/dashboard", func(r *evo.Request) {
		r.View(nil, "template.default")
	})
}

// Permissions setup permissions of app
func (App) Permissions() []evo.Permission {
	return []evo.Permission{
		{Title: "Login to admin", CodeName: "login", Description: "Able login to admin panel"},
		{Title: "Edit own dashboard", CodeName: "dashboard", Description: "Able edit own dashboard otherwise it inherit dashboard assigned by admin"},
	}
}

// Menus setup menus
func (App) Menus() []menu.Menu {
	return []menu.Menu{
		{Title: "Dashboard", Url: "admin/dashboard", Icon: fontawesome.Home},
		{Title: "Parent", Url: "admin/dashboard", Icon: fontawesome.SearchPlus, Children: []menu.Menu{
			{Title: "Child 1", Url: "admin/dashboard", Icon: fontawesome.Image},
			{Title: "Child 2", Url: "admin/dashboard", Icon: fontawesome.Save},
		}},
	}
}

// WhenReady called after setup all apps
func (App) WhenReady() {}
