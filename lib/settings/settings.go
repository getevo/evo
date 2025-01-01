package settings

import (
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/generic"
	"strings"
)

var data = map[string]any{}
var ConfigPath = "./config.yml"

func Get(key string, defaultValue ...any) generic.Value {
	key = strings.ToUpper(key)
	if v, ok := data[key]; ok {
		return generic.Parse(v)
	}
	if len(defaultValue) > 0 {
		return generic.Parse(defaultValue[0])
	}
	return generic.Parse("")
}

func Has(key string) (bool, generic.Value) {
	v, exists := data[key]
	return exists, generic.Parse(v)
}

func All() map[string]any {
	return data
}

func Register(settings ...any) error {
	return nil
}

func Set(key string, value any) error {
	data[key] = generic.Parse(value)
	return nil
}
func SetMulti(in map[string]any) error {
	for key, value := range in {
		data[key] = generic.Parse(value)
	}
	return nil
}

func Init(params ...string) error {
	if args.Get("-c") != "" {
		ConfigPath = args.Get("-c")
	}
	if db.IsEnabled() {
		InitDatabaseSettings()
	}

	err := Reload()
	if err != nil {
		return err
	}
	return nil
}

func Reload() error {
	data = map[string]any{}
	// 1-load environment variables
	LoadEnvVars()

	// 2-load database settings (if enabled)
	if db.IsEnabled() {
		err := LoadDatabaseSettings()
		if err != nil {
			return err
		}
	}

	// 3- load yml
	err := LoadYAMLSettings(ConfigPath)
	if err != nil {
		return err
	}

	// 4- override args
	LoadOSArgs()

	return nil
}
