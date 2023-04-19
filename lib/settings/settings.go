package settings

import (
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/settings/yml"
)

var drivers = []Interface{&yml.Yaml{}}
var defaultDriver Interface = &proxy{}

type Interface interface {
	// Name returns driver name
	Name() string
	Get(key string) generic.Value
	Has(key string) (bool, generic.Value)
	All() map[string]generic.Value
	Set(key string, value interface{}) error
	SetMulti(data map[string]interface{}) error
	Register(settings ...interface{}) error
	Init(params ...string) error
}

type Setting struct {
	Domain      string `gorm:"column:domain;primaryKey" json:"domain"`
	Name        string `gorm:"column:name;primaryKey" json:"name"`
	Title       string `gorm:"column:title" json:"title"`
	Description string `gorm:"column:description" json:"description"`
	Value       string `gorm:"column:value" json:"value"`
	Type        string `gorm:"column:type" json:"type"`
	Params      string `gorm:"column:params" json:"params"`
	ReadOnly    bool   `gorm:"column:read_only" json:"read_only"`
	Visible     bool   `gorm:"column:visible" json:"visible"`
}

type SettingDomain struct {
	DomainID    int    `gorm:"column:domain_id;primaryKey" json:"domain_id"`
	Title       string `gorm:"column:title" json:"title"`
	Description string `gorm:"column:description" json:"description"`
	Domain      string `gorm:"column:domain" json:"domain"`
	ReadOnly    bool   `gorm:"column:read_only" json:"read_only"`
	Visible     bool   `gorm:"column:visible" json:"visible"`
}

func SetDefaultDriver(driver Interface) {
	AddDriver(driver)
	defaultDriver = driver
}

func DriverName() string {
	return defaultDriver.Name()
}

func Drivers() map[string]Interface {
	var list = map[string]Interface{}
	for idx, item := range drivers {
		list[item.Name()] = drivers[idx]
	}
	return list
}

func Driver(driver string) (Interface, bool) {

	for idx, item := range drivers {
		if item.Name() == driver {
			return drivers[idx], true
		}
	}
	return nil, false
}

func Use(driver string) Interface {
	for idx, item := range drivers {
		if item.Name() == driver {
			return drivers[idx]
		}
	}
	return nil
}

func AddDriver(driver Interface) {
	if _, ok := Driver(driver.Name()); !ok {
		drivers = append(drivers, driver)
		var err = driver.Init()
		if err != nil {
			log.Fatal("unable to initiate pub/sub driver", "name", driver.Name(), "error", err)
		}
	}
	if defaultDriver == nil {
		defaultDriver = driver
	}
}

func Get(key string) generic.Value {
	return defaultDriver.Get(key)
}

func Has(key string) (bool, generic.Value) {
	return defaultDriver.Has(key)
}

func All() map[string]generic.Value {
	return defaultDriver.All()
}

func Set(key string, value interface{}) error {
	return defaultDriver.Set(key, value)
}

func SetMulti(data map[string]interface{}) error {
	return defaultDriver.SetMulti(data)
}

func Register(settings ...interface{}) {
	defaultDriver.Register(settings...)
}

func Init(params ...string) error {
	return defaultDriver.Init(params...)
}

func (SettingDomain) TableName() string {
	return "settings_domain"
}

func (Setting) TableName() string {
	return "settings"
}
