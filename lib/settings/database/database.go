package database

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/db"

	"github.com/getevo/evo/v2/lib/generic"
	"gorm.io/gorm/clause"
	"strings"
	"sync"
)

var Driver = &Database{}
var domains = map[string]SettingDomain{}

type Database struct {
	mu   sync.Mutex
	data map[string]map[string]generic.Value
}

func (config *Database) Name() string {
	return "database"
}

func (config *Database) Get(key string) generic.Value {
	key = strings.ToUpper(key)
	var chunks = strings.SplitN(key, ".", 2)
	if len(chunks) != 2 {
		return generic.Parse("")
	}
	if domain, ok := config.data[chunks[0]]; ok {
		if v, exists := domain[chunks[1]]; exists {
			return v
		}
	}
	return generic.Parse("")
}
func (config *Database) Has(key string) (bool, generic.Value) {
	key = strings.ToUpper(key)
	var chunks = strings.SplitN(key, ".", 2)
	if len(chunks) != 2 {
		return false, generic.Parse("")
	}
	if domain, ok := config.data[chunks[0]]; ok {
		if v, exists := domain[chunks[1]]; exists {
			return true, v
		}
	}
	return false, generic.Parse("")
}
func (config *Database) All() map[string]generic.Value {
	var m = map[string]generic.Value{}
	for domain, inner := range config.data {
		for name, value := range inner {
			m[domain+"."+name] = value
		}
	}
	return m
}
func (config *Database) Set(key string, value interface{}) error {
	key = strings.ToUpper(key)
	var chunks = strings.SplitN(key, ".", 2)
	if len(chunks) == 2 {
		db.Where("domain = ? AND name = ?", chunks[0], chunks[1]).Model(Setting{}).Update("value", value)
	}
	return nil
}
func (config *Database) SetMulti(data map[string]interface{}) error {
	for key, value := range data {
		key = strings.ToUpper(key)
		var chunks = strings.SplitN(key, ".", 2)
		if len(chunks) == 2 {
			db.Where("domain = ? AND name = ?", chunks[0], chunks[1]).Model(Setting{}).Update("value", value)
		}
	}
	return nil
}
func (config *Database) Register(settings ...interface{}) error {
	for _, s := range settings {
		var v = generic.Parse(s)

		if v.Is("settings.Setting") {
			var setting = Setting{}
			var err = v.Cast(&setting)
			if err != nil {
				return err
			}
			if ok, _ := config.Has(setting.Domain + "." + setting.Name); !ok {
				config.Set(setting.Domain+"."+setting.Name, setting.Value)
				db.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&setting)
				if _, exists := domains[setting.Domain]; !exists {
					domain := SettingDomain{
						Title:       setting.Domain,
						Domain:      setting.Domain,
						Description: "",
						ReadOnly:    false,
						Visible:     true,
					}
					db.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&domain)
				}
			}
		} else if v.Is("settings.SettingDomain") {
			fmt.Println(v.Input)
			var domain = SettingDomain{}
			var err = v.Cast(&domain)
			if err != nil {
				return err
			}

			if _, exists := domains[domain.Domain]; !exists {
				db.Debug().Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&domain)
			}
		}

	}
	return nil
}
func (config *Database) Init(params ...string) error {
	config.mu.Lock()
	var items []Setting
	if args.Exists("-migrate") {
		db.Unscoped().AutoMigrate(&Setting{}, &SettingDomain{})
	}

	config.data = make(map[string]map[string]generic.Value)

	db.Debug().Find(&items)
	for _, item := range items {
		item.Domain = strings.ToUpper(item.Domain)
		item.Name = strings.ToUpper(item.Name)
		if _, ok := config.data[item.Domain]; !ok {
			config.data[item.Domain] = map[string]generic.Value{}
		}
		config.data[item.Domain][item.Name] = generic.Parse(item.Value)
	}
	var list []SettingDomain
	db.Debug().Find(&list)
	for idx, item := range list {
		domains[item.Domain] = list[idx]
	}
	config.mu.Unlock()
	return nil
}
