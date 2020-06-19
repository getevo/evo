package evo

import (
	"github.com/getevo/evo/lib/gpath"
	"github.com/getevo/evo/menu"
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

var Docs = []DocApp{}

type DocApp struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Author      string       `json:"author"`
	Namespace   string       `json:"namespace"`
	Path        string       `json:"path"`
	Models      []*DocModel  `json:"models"`
	APIs        []*DocAPI    `json:"apis"`
	Permissions []Permission `json:"permissions"`
	Menus       []menu.Menu  `json:"menus"`
	Package     string       `json:"package"`
}

type DocAPI struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Describe    []string `json:"describe"`
	Body        string   `json:"body"`
	Return      string   `json:"return"`
	Required    string   `json:"required"`
	Method      string   `json:"method"`
	URL         string   `json:"url"`
	Package     string   `json:"package"`
	Path        string   `json:"path"`
}

type DocModel struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Methods     []string `json:"methods"`
	Fields      []string `json:"fields"`
	Package     string   `json:"package"`
	Path        string   `json:"path"`
}

func NewDoc(app App) {
	typ := reflect.TypeOf(app)
	path := build.Default.GOPATH + "/src/" + typ.PkgPath()
	doc := DocApp{
		APIs: []*DocAPI{},
	}

	doc.Menus = app.Menus()
	doc.Permissions = app.Permissions()
	doc.Namespace = strings.Split(reflect.ValueOf(app).Type().String(), ".")[0]

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			content := strings.Split(string(gpath.SafeFileContent(path)), "\n")
			parseDoc(&doc, content, typ.PkgPath())
		}
		return nil
	})

	Docs = append(Docs, doc)

}

func parseDoc(app *DocApp, content []string, path string) {
	var parser = regexp.MustCompile(`\/\/\s+@doc\s(\w+)\s+(.+)`)
	var model *DocModel
	var api *DocAPI
	var meta = map[string]string{}
	var typ = ""
	for ln, line := range content {
		line = strings.TrimSpace(line)
		matches := parser.FindAllStringSubmatch(line, 1)
		if len(matches) == 1 {
			cmd := matches[0][1]
			value := matches[0][2]
			if cmd == "type" {
				typ = value
				if typ == "api" {
					api = &DocAPI{
						Describe: []string{},
					}
					app.APIs = append(app.APIs, api)
				}
				if typ == "model" {
					model = &DocModel{}
					app.Models = append(app.Models, model)
				}
				continue
			}
			if cmd == "include" {
				includeDoc(app, value)
				continue
			}
			if typ == "api" && api != nil {
				api.setReflectValue(cmd, value)
			}
			if typ == "model" && model != nil {
				model.setReflectValue(cmd, value)
			}
			if typ == "app" {
				app.setReflectValue(cmd, value)
			}
			if typ == "meta" {
				meta[cmd] = value
			}
		} else {
			if strings.TrimSpace(line) == "" {
				continue
			}

			if ln < 10 && strings.HasPrefix(strings.TrimSpace(line), "package") {
				parts := strings.Fields(strings.TrimSpace(line))
				if len(parts) == 2 {
					meta["package"] = parts[1]
				}
				continue
			}
			if typ == "api" && api != nil {
				api.Parse(content, ln, meta)
			}
			if typ == "model" && model != nil {
				model.Parse(content, ln, meta)
				if v, ok := meta["package"]; ok {
					model.Type = v + "." + model.Name
				} else {
					model.Type = app.Namespace + "." + model.Name
				}

			}
			typ = ""
		}
	}

}

func includeDoc(doc *DocApp, path string) {
	p := build.Default.GOPATH + "/src/" + path
	if gpath.IsDir(p) {
		filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
				content := strings.Split(string(gpath.SafeFileContent(path)), "\n")
				parseDoc(doc, content, path)
			}
			return nil
		})
	} else if gpath.IsFileExist(p) {
		content := strings.Split(string(gpath.SafeFileContent(p)), "\n")
		parseDoc(doc, content, p)
	}

}

func (obj *DocAPI) setReflectValue(key string, value string) {
	if key == "describe" {
		obj.Describe = append(obj.Describe, value)
		return
	}
	el := reflect.ValueOf(obj).Elem()
	for i := 0; i < el.Type().NumField(); i++ {
		if el.Type().Field(i).Tag.Get("json") == key {
			el.Field(i).SetString(value)
			return
		}
	}

}

func (obj *DocAPI) Parse(content []string, i int, meta map[string]string) {
	re := regexp.MustCompile(`\w+\.(\w+)\(\"(.+)\"`)
	res := re.FindAllStringSubmatch(content[i], 1)
	if len(res) == 1 {
		obj.Method = strings.ToUpper(res[0][1])
		obj.URL = res[0][2]
		if v, ok := meta["prefix"]; ok {
			obj.URL = v + obj.URL
		}
	}
}

func (obj *DocModel) setReflectValue(key string, value string) {
	el := reflect.ValueOf(obj).Elem()
	for i := 0; i < el.Type().NumField(); i++ {
		if el.Type().Field(i).Tag.Get("json") == key {
			el.Field(i).SetString(value)
			return
		}
	}

}

func (obj *DocModel) Parse(content []string, i int, meta map[string]string) {
	parts := strings.Fields(content[i])
	if len(parts) >= 3 {
		obj.Name = parts[1]
	} else {
		return
	}
	for {
		i++
		content[i] = strings.TrimSpace(content[i])
		fields := strings.Fields(content[i])
		if strings.HasPrefix(content[i], "\\") {
			continue
		}
		if len(fields) == 2 {
			obj.Fields = append(obj.Fields, content[i])
		} else if len(fields) > 2 {
			obj.Fields = append(obj.Fields, content[i])
		}

		if i == len(content) {
			return
		}
		if len(content[i]) > 0 && content[i][0] == '}' {
			return
		}
	}
}

func (obj *DocApp) setReflectValue(key string, value string) {
	el := reflect.ValueOf(obj).Elem()
	for i := 0; i < el.Type().NumField(); i++ {
		if el.Type().Field(i).Tag.Get("json") == key {
			el.Field(i).SetString(value)
			return
		}
	}
}
