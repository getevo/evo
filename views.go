package evo

import (
	"fmt"
	"github.com/iesreza/jet/v8"
	"io"
	"reflect"
)

//TODO: Concurrent View Pool
type views map[string]*jet.Set

var viewList = views{}
var viewFnApplied = map[string]bool{}

//RegisterView register views of given path
func RegisterView(prefix, path string, safeWriter ...jet.SafeWriter) *jet.Set {
	if len(safeWriter) == 0 {
		viewList[prefix] = jet.NewSet(jet.NewOSFileSystemLoader(path), jet.WithSafeWriter(func(writer io.Writer, bytes []byte) {
			writer.Write(bytes)
		}))

	} else {
		viewList[prefix] = jet.NewSet(jet.NewOSFileSystemLoader(path), jet.WithSafeWriter(safeWriter[0]))
	}
	viewList[prefix].Development(config.Server.Debug)
	applyViewFunction(prefix)
	return viewList[prefix]
}

//GetView return view of given environment
func GetView(prefix, name string) (*jet.Template, error) {
	if t, ok := viewList[prefix]; ok {
		return t.GetTemplate(name)
	}

	return nil, fmt.Errorf("template prefix \"%s\" not found", prefix)

}

var globalFunctions = map[string]func(arguments jet.Arguments) reflect.Value{}

func RegisterViewFunction(name string, fn func(arguments jet.Arguments) reflect.Value) {
	globalFunctions[name] = fn
}

func applyViewFunction(prefix string) {
	for name, fn := range globalFunctions {
		if ok, _ := viewFnApplied[prefix+"#"+name]; !ok {
			viewList[prefix].AddGlobalFunc(name, fn)
			viewFnApplied[prefix+"#"+name] = true
		}
	}
}

var viewGlobalParams = map[string]interface{}{}

func AddGlobalViewParam(key string, value interface{}) {
	viewGlobalParams[key] = value
}
