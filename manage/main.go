package main

import (
	"bytes"
	"fmt"
	"github.com/getevo/evo/lib/date"
	"github.com/getevo/evo/lib/gpath"
	"github.com/getevo/evo/lib/text"
	"github.com/getevo/evo/manage/tools/gaper"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	git           = "https://github.com/getevo/evo.git"
	path          = os.Getenv("GOPATH") + "/src/github.com/getevo/evo/apps/standard"
	command       = ""
	workingdir, _ = os.Getwd()
	app           = filepath.Base(workingdir)
	subcommand    = ""
)

func main() {
	if len(os.Args) > 1 {
		command = os.Args[1]
		if command == "dev" {
			cfg := gaper.Config{
				BinName:      app,
				BuildPath:    workingdir,
				PollInterval: gaper.DefaultPoolInterval,
			}
			chOSSiginal := make(chan os.Signal, 2)
			if err := gaper.Run(&cfg, chOSSiginal); err != nil {
				panic(err)
			}
		}

		if command == "create" {
			data, err, h := Expect("type=>app|model|docker", "name")
			if err == nil {
				create(data)
				return
			}
			fmt.Println(err)
			fmt.Println("Usage: " + h)
		}
	}
	help()
}

func create(data map[string]string) {
	data = parseVars(data)

	if data["type"] == "app" {
		gpath.MakePath(workingdir + "/" + data["name"])
		b, err := gpath.ReadFile(path + "/app.go")
		if err != nil {
			panic(err)
		}
		content := strings.Replace(string(b), "__PACKAGE__", data["name"], -1)

		content = render(content, data)
		fmt.Println(content)
		f, err := gpath.Open(workingdir + "/" + data["name"] + "/app.go")
		if err != nil {
			panic(err)
		}
		f.WriteString(content)

	}

	if data["type"] == "model" {
		gpath.MakePath(workingdir + "/" + data["name"])
		b, err := gpath.ReadFile(path + "/model.go")
		if err != nil {
			panic(err)
		}
		content := strings.Replace(string(b), "__PACKAGE__", data["name"], -1)

		content = render(content, data)
		fmt.Println(content)
		f, err := gpath.Open(workingdir + "/" + data["name"] + "/app.go")
		if err != nil {
			panic(err)
		}
		f.WriteString(content)

	}
}

func parseVars(data map[string]string) map[string]string {
	data["date"] = date.Now().Format("2006-01-02 15:04:05")
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	data["user"] = user.Name + " - " + user.Username
	return data
}

func help() {

}

func render(s string, data map[string]string) string {
	var tpl bytes.Buffer
	t := template.New("action")
	t, err := t.Parse(s)
	if err != nil {
		panic(err)
	}
	if err := t.Execute(&tpl, data); err != nil {
		panic(err)
	}

	return tpl.String()
}

func Expect(params ...string) (map[string]string, error, string) {
	var t int
	args := os.Args[2:]
	res := map[string]string{}

	for _, item := range params {
		t = 0
		if item[0] == '-' {
			t = 1
			if item[1] == '-' {
				t = 2
			}
		}

		if (t == 1 || t == 2) && len(args) < 2 {
			return res, fmt.Errorf("e1: %s is not satisfied", item), usage(params)
		}
		if t == 0 && strings.Contains(item, "=>") {
			opt := text.ParseWildCard(item, `*=>*`)
			if len(args) == 0 {
				return res, fmt.Errorf("e2: %s is not satisfied", opt[0]), usage(params)
			}

			chunks := strings.Split(opt[1], "|")
			found := false
			if len(args) == 0 {
				return res, fmt.Errorf("e6: %s is not satisfied", opt[0]), usage(params)
			}
			for _, c := range chunks {

				if args[0] == c {
					found = true
					args = args[1:]
					res[opt[0]] = c
					break
				}
			}
			if !found {
				return res, fmt.Errorf("e3: %s is not satisfied", opt[0]), usage(params)
			}
			continue
		} else if t == 0 {
			if len(args) == 0 {
				return res, fmt.Errorf("e7: %s is not satisfied", item), usage(params)
			}
			res[item] = args[0]
			args = args[1:]
			continue
		}

		found := false
		for index, arg := range args {
			if arg == item {
				if len(args) < index+1 {
					return res, fmt.Errorf("e4: %s is not satisfied", item), usage(params)
				}
				res[arg] = args[index+1]
				found = true
				continue
			}
		}
		if !found && t == 1 {
			return res, fmt.Errorf("e5: %s is not satisfied", item), usage(params)
		}

	}

	return res, nil, ""

}

func usage(params []string) string {
	var res = "io " + command
	var t int
	for _, item := range params {
		t = 0
		if item[0] == '-' {
			t = 1
			if item[1] == '-' {
				t = 2
			}
		}
		if t == 0 && strings.Contains(item, "=>") {
			opt := text.ParseWildCard(item, `*=>*`)
			res += " " + opt[1]
			continue
		} else if t == 0 {
			res += " " + item
			continue
		}

		if t == 1 {
			res += " " + item + " " + strings.Trim(item, "-")
			continue
		}

		if t == 2 {
			res += " [" + item + " " + strings.Trim(item, "-") + "]"
			continue
		}
	}
	return res
}
