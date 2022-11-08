package settings

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/lib/concurrent"
	"github.com/getevo/evo/lib/fontawesome"
	"github.com/getevo/evo/menu"
	"github.com/iesreza/jet/v8"
	"gorm.io/gorm"
	"reflect"
)

var controller Controller
var settings = concurrent.Map{}
var db *gorm.DB
var config *evo.Configuration
var views *jet.Set
var Path string

// App settings app struct
type App struct{}

var initiated = false

// Register register the auth in io apps
func Register(v ...interface{}) {
	if len(v) == 0 {
		evo.Register(App{})
		return
	}
	if initiated == false {
		Register()
	}
	var title string
	var object interface{}
	for _, item := range v {
		ref := reflect.ValueOf(item)
		switch ref.Kind() {
		case reflect.String:
			title = item.(string)
			break
		case reflect.Ptr:
			object = item
			break
		}
	}

	if title != "" && object != nil {
		controller.set(title, object)
	}

}

// Register settings app
func (App) Register() {
	fmt.Println("Settings Registered")
	settings.Init()
	db = evo.GetDBO()
	config = evo.GetConfig()
	if config.Database.Enabled == false {
		panic("Auth App require database to be enabled. solution: enable database at config.yml")
	}
	Path = evo.GuessAsset(App{})
	views = evo.RegisterView("settings", Path+"/views")
	if evo.Arg.Migrate {
		db.AutoMigrate(&Settings{})
	}
}

// Router setup routers
func (App) Router() {
	controller := Controller{}
	evo.Get("admin/settings", controller.view)
	evo.Post("admin/settings/:name", controller.save)
	evo.Post("admin/settings/reset/:name", controller.reset)
}

// Permissions setup permissions of app
func (App) Permissions() []evo.Permission {
	return []evo.Permission{
		{Title: "Access Settings", CodeName: "view", Description: "Access list to view list of settings"},
		{Title: "Modify Settings", CodeName: "modify", Description: "Modify Settings"},
	}
}

// Menus setup menus
func (App) Menus() []menu.Menu {
	return []menu.Menu{
		{Title: "Settings", Url: "admin/settings", Permission: "settings.view", Icon: fontawesome.Cog},
	}
}

// WhenReady called after setup all apps
func (App) WhenReady() {}

func (App) Pack() {}
