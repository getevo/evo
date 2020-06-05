package evo

import (
	"github.com/getevo/evo/lib/ref"
	"github.com/getevo/evo/menu"
	"github.com/getevo/evo/user"
)

type App interface {
	Register()
	Router()
	WhenReady()
	Permissions() []user.Permission
	Menus() []menu.Menu
}

var onReady = []func(){}
var apps = map[string]interface{}{}
var AppMenus = []menu.Menu{}

// Register register app to use by IO
func Register(app App) {
	name := ref.Parse(app).Package

	//app already exist
	if _, ok := apps[name]; ok {
		return
	}

	apps[name] = app
	app.Register()
	app.Router()
	permissions := user.Permissions(app.Permissions())
	permissions.Sync(name)
	n := app.Menus()
	AppMenus = append(AppMenus, n...)

	onReady = append(onReady, app.WhenReady)
}

// GetRegisteredApps return list of registered apps
func GetRegisteredApps() map[string]interface{} {
	return apps
}
