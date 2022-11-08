package bible

import (
	"encoding/json"
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/menu"
)

func Register() {
	evo.Register(App{})
}

type App struct{}

var config *evo.Configuration
var views *jet.Set
var Path string

// Register the bible
func (App) Register() {
	fmt.Println("Bible Registered")
	config = evo.GetConfig()
	Path = evo.GuessAsset(App{})
	views = evo.RegisterView("bible", Path+"/views")

}

// WhenReady called after setup all apps
func (App) WhenReady() {
	views.AddGlobal("title", config.App.Name)
}

// Router setup routers
func (App) Router() {
	evo.Static("/assets", Path+"/assets")
	evo.Get("/bible", func(r *evo.Request) {
		data, _ := json.Marshal(evo.Docs)
		r.View(string(data), "bible.default")
	})
}

// Permissions setup permissions of app
func (App) Permissions() []evo.Permission { return []evo.Permission{} }

// Menus setup menus
func (App) Menus() []menu.Menu {
	return []menu.Menu{}
}

func (App) Pack() {}
