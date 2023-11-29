package yml

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/dot"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/gpath"
	"gopkg.in/yaml.v3"
)

var Driver = &Yaml{}

type Yaml struct {
	data    map[string]interface{}
	path    string
	writeFn func()
}

func (config *Yaml) Name() string {
	return "yml"
}

func (config *Yaml) Get(key string) generic.Value {
	var data, err = dot.Get(config.data, key)
	if err != nil {
		return generic.Parse("")
	}
	var value = generic.Parse(data)
	return value
}
func (config *Yaml) Has(key string) (bool, generic.Value) {
	var data, err = dot.Get(config.data, key)
	var value = generic.Parse(data)
	if err != nil || data == nil {
		return false, generic.Parse(nil)
	}
	return true, value
}
func (config *Yaml) All() map[string]generic.Value {
	return map[string]generic.Value{}
}
func (config *Yaml) Set(key string, value interface{}) error {
	dot.Set(config.data, key, value)
	return config.write()
}
func (config *Yaml) SetMulti(data map[string]interface{}) error {
	for key, value := range data {
		dot.Set(&config.data, key, value)
	}

	return config.write()
}
func (config *Yaml) Register(settings ...interface{}) error {
	if config.data == nil {
		config.data = map[string]interface{}{}
	}
	for _, s := range settings {
		var v = generic.Parse(s)
		if !v.Is("settings.Setting") {
			return fmt.Errorf("invalid settings")
		}
		if ok, _ := config.Has(v.Prop("Domain").String() + "." + v.Prop("Name").String()); !ok {
			config.Set(v.Prop("Domain").String()+"."+v.Prop("Name").String(), v.Prop("Value").String())
		}

	}
	return nil
}
func (config *Yaml) Init(params ...string) error {
	config.data = map[string]interface{}{}
	if len(params) != 0 {
		for _, path := range params {
			bytes, err := gpath.ReadFile(path)
			if err == nil {
				config.path = path
				err = yaml.Unmarshal(bytes, &config.data)
				return err
			}
		}
	} else {
		var path = args.Get("-c")
		if path == "" {
			path = "./config.yml"
		}
		bytes, err := gpath.ReadFile(path)
		if err == nil {
			config.path = path
			err = yaml.Unmarshal(bytes, &config.data)
			return err
		}
	}

	return fmt.Errorf("config file not found")
}

func (config *Yaml) write() error {
	var bytes, err = yaml.Marshal(config.data)
	if err != nil {
		return err
	}
	file, err := gpath.Open(config.path)
	if err != nil {
		return err
	}
	err = file.Write(bytes)
	if err != nil {
		return err
	}
	file.Close()

	return nil
}
