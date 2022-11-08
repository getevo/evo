package test

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/lib/gpath"
	"github.com/getevo/evo/menu"
	"github.com/iesreza/jet/v8"
	"gorm.io/gorm"
)

func Register() {
	fmt.Println("Test Registered")
	evo.Register(App{})
}

var db *gorm.DB
var Path string

type App struct{}

var views *jet.Set

func (App) Register() {
	//Require auth
	db = evo.GetDBO()
	Path = gpath.Parent(gpath.WorkingDir()) + "/apps/test"
	views = evo.RegisterView("test", Path+"/views")
}

func (App) Router() {
	evo.Get("/admin/list", FilterViewController)
	evo.Get("test1", func(request *evo.Request) {
		u := evo.User{}
		db.Find(&u)
		request.WriteResponse(u)
	})
}

func (App) Permissions() []evo.Permission {
	return []evo.Permission{}
}

func (App) Menus() []menu.Menu {
	return []menu.Menu{}
}
func (App) WhenReady() {
	if evo.Arg.Migrate {
		db.AutoMigrate(MyModel{}, MyGroup{})
	}

	/*	db.Debug().Create(&MyGroup{
			Name:"Group 1",
		})
		db.Debug().Create(&MyGroup{
			Name:"Group 2",
		})
		db.Debug().Create(&MyGroup{
			Name:"Group 3",
		})*/
}

func (App) Pack() {}
