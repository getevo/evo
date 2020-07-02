package evo

import (
	"github.com/getevo/evo/lib/gpath"
	"github.com/getevo/evo/lib/log"
	"github.com/getevo/evo/lib/ref"
	"github.com/getevo/evo/menu"
	"go/build"
	"reflect"
	"strings"
)

type App interface {
	Register()
	Router()
	WhenReady()
	Permissions() []Permission
	Pack()
	Menus() []menu.Menu
}

var onReady = []func(){}
var apps = map[string]interface{}{}
var AppMenus = []menu.Menu{}

// Register register app to use by EVO
func Register(app App) {
	name := ref.Parse(app).Package

	//app already exist
	if _, ok := apps[name]; ok {
		return
	}

	apps[name] = app
	app.Register()
	app.Router()
	permissions := Permissions(app.Permissions())
	permissions.Sync(name)
	n := app.Menus()
	AppMenus = append(AppMenus, n...)

	onReady = append(onReady, app.WhenReady)
	if config.Server.Debug {
		NewDoc(app)
	}
	if Arg.Pack {
		app.Pack()
	}
}

// GetRegisteredApps return list of registered apps
func GetRegisteredApps() map[string]interface{} {
	return apps
}

func GuessAsset(app App) string {
	src := ""
	pack := ""
	t := reflect.ValueOf(app)
	if t.Kind() == reflect.Ptr {
		src = t.Elem().Type().PkgPath()
		pack = t.Elem().Type().String()
	} else {
		src = t.Type().PkgPath()
		pack = t.Type().String()
	}
	//
	pack = strings.Replace(pack, ".", "/", 1)
	pack = "/bundle/" + gpath.Parent(pack)
	pack = strings.Trim(pack, "/")

	if gpath.IsDirExist(gpath.WorkingDir() + "/" + pack) {
		log.Info("Load local bundle at " + gpath.WorkingDir() + "/" + pack)
		return gpath.WorkingDir() + "/" + pack
	}

	if gpath.IsDirExist(build.Default.GOPATH + "/src/" + src) {
		log.Info("Load bundle from " + build.Default.GOPATH + "/src/" + src)
		return build.Default.GOPATH + "/src/" + src
	}

	panic("Unable to guess asset path for " + pack)
	return ""
}
