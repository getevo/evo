package application

import (
	"github.com/getevo/evo/v2/lib/log"
	"reflect"
	"sort"
	"strings"
)

const (
	HIGHEST Priority = 0
	HIGH    Priority = 1
	DEFAULT Priority = 5
	LOW     Priority = 6
	LOWEST  Priority = 7
)

type Priority int

type Application interface {
	Register() error
	Router() error
	WhenReady() error
	Name() string
}

type PriorityInterface interface {
	Priority() Priority
}

type ReloadInterface interface {
	Reload() error
}

type App struct {
	apps []Application
}

var Instance *App

func GetInstance() *App {
	if Instance == nil {
		Instance = &App{}
	}
	return Instance
}

func ReloadAll() {
	for _, app := range GetInstance().apps {
		if r, ok := app.(ReloadInterface); ok {
			if err := r.Reload(); err != nil {
				log.Criticalf("Can't reload application %s: %v", app.Name(), err)
			}
		}
	}
}

func (a *App) Register(applications ...Application) *App {
	a.apps = append(a.apps, applications...)
	return a
}

func (a *App) Run() *App {
	//Sort applications by priority
	sort.Slice(a.apps, func(i, j int) bool {
		return getPriority(a.apps[i]) < getPriority(a.apps[j])
	})

	for _, app := range a.apps {
		if err := app.Register(); err != nil {
			log.Fatalf("Can't start application Register() %s: %v", app.Name(), err)
		}

		var ref = reflect.ValueOf(app)
		typ := reflect.TypeOf(app)
		for i := 0; i < typ.NumMethod(); i++ {
			method := typ.Method(i)
			if strings.HasPrefix(method.Name, "OnRegister") {
				ref.Method(i).Call([]reflect.Value{})
			}
		}
		if err := app.Router(); err != nil {
			log.Fatalf("Can't start application Router() %s: %v", app.Name(), err)
		}
	}
	for _, app := range a.apps {
		if err := app.WhenReady(); err != nil {
			log.Fatalf("Can't start application WhenReady() %s: %v", app.Name(), err)
		}
	}

	return a
}

func (a *App) GetApps() []Application {
	return a.apps
}

func getPriority(input interface{}) Priority {
	if v, ok := input.(PriorityInterface); ok {
		return v.Priority()
	}
	return DEFAULT
}
