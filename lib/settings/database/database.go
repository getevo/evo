package database

import (
	"errors"
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/settings"

	"strings"
	"sync"

	"github.com/getevo/evo/v2/lib/generic"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var Driver = &Database{}
var dbVersion uint
var domains = map[string]SettingDomain{}

type Database struct {
	mu    sync.Mutex
	data0 map[string]map[string]generic.Value
	data1 map[string]generic.Value
}

func (config *Database) Name() string {
	return "database"
}

func (config *Database) Get(key string) generic.Value {
	key = strings.ToUpper(key)
	switch dbVersion {
	case 0:
		var chunks = strings.SplitN(key, ".", 2)
		if len(chunks) != 2 {
			return generic.Parse("")
		}
		if domain, ok := config.data0[chunks[0]]; ok {
			if v, exists := domain[chunks[1]]; exists {
				return v
			}
		}
		break
	case 1:
		if v, exists := config.data1[key]; exists {
			return v
		}
	}
	return generic.Parse("")
}
func (config *Database) Has(key string) (bool, generic.Value) {
	key = strings.ToUpper(key)
	switch dbVersion {
	case 0:
		var chunks = strings.SplitN(key, ".", 2)
		if len(chunks) != 2 {
			return false, generic.Parse("")
		}
		if domain, ok := config.data0[chunks[0]]; ok {
			if v, exists := domain[chunks[1]]; exists {
				return true, v
			}
		}
		break
	case 1:
		if v, exists := config.data1[key]; exists {
			return true, v
		}
	}
	return false, generic.Parse("")
}
func (config *Database) All() map[string]generic.Value {
	switch dbVersion {
	case 0:
		var m = map[string]generic.Value{}
		for domain, inner := range config.data0 {
			for name, value := range inner {
				m[domain+"."+name] = value
			}
		}
		return m
	case 1:
		return config.data1
	}
	return nil
}

func (config *Database) Set(key string, value any) error {
	key = strings.ToUpper(key)

	switch dbVersion {
	case 0:
		var chunks = strings.SplitN(key, ".", 2)
		if len(chunks) == 2 {
			db.Where("domain = ? AND name = ?", chunks[0], chunks[1]).Model(Setting{}).Update("value", value)
		}
		break
	case 1:
		// Not implemented
		break
	}
	return nil
}

func (config *Database) SetMulti(data map[string]any) error {
	for key, value := range data {
		key = strings.ToUpper(key)
		var chunks = strings.SplitN(key, ".", 2)
		if len(chunks) == 2 {
			db.Where("domain = ? AND name = ?", chunks[0], chunks[1]).Model(Setting{}).Update("value", value)
		}
	}
	return nil
}

func (config *Database) Register(sets ...any) error {
	switch dbVersion {
	case 0:
		for _, s := range sets {
			var v = generic.Parse(s)

			if v.Is("settings.Setting") {
				var setting = SettingLegacy{}
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
				var domain = SettingDomainLegacy{}
				var err = v.Cast(&domain)
				if err != nil {
					return err
				}

				if _, exists := domains[domain.Domain]; !exists {
					db.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&domain)
				}
			}

		}
		break
	case 1:
		for _, s := range sets {
			var v = generic.Parse(s)
			if v.Is("settings.SettingDomain") {
				var domain SettingDomain
				var err = v.Cast(&domain)
				if err != nil {
					return err
				}

				// This limits the domains to 2 for a parameter
				var portions = strings.SplitN(domain.Domain, ".", 2)
				var parentDomain SettingDomain
				var grandfatherDomainId uint
				var skip = false
				switch len(portions) {
				case 1:
					err := db.Debug().Model(&SettingDomain{}).Where("(parent_domain = 0 OR parent_domain IS NULL) AND domain = ?", portions[0]).First(&parentDomain).Error
					if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
						return err
					}
					domain.Domain = portions[0]
					parentDomain = domain
					grandfatherDomainId = parentDomain.ID
					skip = true
					break
				case 2:
					err := db.Debug().Model(&SettingDomain{}).Preload("ChildrenDomains", "Domain = ?", portions[1]).Preload("ChildrenDomains.Parameters").Where("(parent_domain = 0 or parent_domain IS NULL) AND domain = ?", portions[0]).First(&parentDomain).Error
					if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
						return err
					}
					domain.Domain = portions[1]
					parentDomain.ChildrenDomains = append(parentDomain.ChildrenDomains, domain)
					parentDomain.Domain = portions[0]
					grandfatherDomainId = 0
					skip = true
					break
				}
				if parentDomain.ParentDomain != nil && *parentDomain.ParentDomain == 0 {
					parentDomain.ParentDomain = nil
				}
				if !skip {
					err = db.Debug().Clauses(clause.Insert{Modifier: "IGNORE"}).Where("domain = ? AND parent_domain = ?", parentDomain.Domain, grandfatherDomainId).Save(&parentDomain).Error
					if err != nil {
						return err
					}
				}

			}
		}
		for _, s := range sets {
			var v = generic.Parse(s)
			if v.Is("settings.Setting") {
				var setting = settings.Setting{}
				var err = v.Cast(&setting)
				if err != nil {
					return err
				}

				if setting.Value != "" {
					config.data1[setting.Domain+"."+setting.Name] = generic.Parse(setting.Value)
				}

				// First let's check if the domain and subdomain exists and create them if they dont
				// This limits the domains to 2 for a parameter
				var portions = strings.SplitN(setting.Domain, ".", 2)
				var parentDomain SettingDomain
				var toSave = Setting{
					Name:        setting.Name,
					Description: setting.Description,
					Title:       setting.Title,
					Value:       setting.Value,
					ReadOnly:    setting.ReadOnly,
					Visible:     setting.ReadOnly,
					Type:        setting.Type,
					Params:      setting.Params,
				}

				switch len(portions) {
				case 1:
					err := db.Model(&SettingDomain{}).Where("(parent_domain = 0 OR parent_domain IS NULL) AND domain = ?", portions[0]).First(&parentDomain).Error
					if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
						return err
					}
					parentDomain.Domain = portions[0]
					if len(parentDomain.Parameters) == 0 {
						parentDomain.Parameters = append(parentDomain.Parameters, toSave)
					} else {
						toSave.ID = parentDomain.Parameters[0].ID
						parentDomain.Parameters[0] = toSave
					}
					break
				case 2:
					err := db.Model(&SettingDomain{}).Preload("ChildrenDomains", "Domain = ?", portions[1]).Preload("ChildrenDomains.Parameters").Where("parent_domain = 0 AND domain = ?", portions[0]).First(&parentDomain).Error
					if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
						return err
					}
					parentDomain.Domain = portions[0]
					if len(parentDomain.ChildrenDomains) == 0 {
						parentDomain.ChildrenDomains = append(parentDomain.ChildrenDomains, SettingDomain{Domain: portions[1], Parameters: []Setting{toSave}})
					} else {
						if len(parentDomain.ChildrenDomains[0].Parameters) == 0 {
							parentDomain.ChildrenDomains[0].Parameters = append(parentDomain.ChildrenDomains[0].Parameters, toSave)
						} else {
							toSave.ID = parentDomain.ChildrenDomains[0].Parameters[0].ID
							parentDomain.ChildrenDomains[0].Parameters[0] = toSave
						}
					}
					break
				}
				if parentDomain.ParentDomain != nil && *parentDomain.ParentDomain != 0 {
					parentDomain.ParentDomain = nil
				}
				// Now we can finally create the setting parameter, subdomain and domain in one go
				err = db.Clauses(clause.Insert{Modifier: "IGNORE"}).Where("parent_domain = 0 AND domain = ?", parentDomain.Domain).Save(&parentDomain).Error
				if err != nil {
					return err
				}

			}
		}
		break
	}

	return nil
}
func (config *Database) Init(params ...string) error {
	// Check if the database structure is the old one or the new one
	if err := db.Model(&SettingDomain{}).Where("parent_domain = 0 OR parent_domain IS NULL").First(&SettingDomain{}).Error; err == nil {
		log.Info("new db version detected")
		dbVersion = 1
	} else {
		log.Info("old db version detected")
		dbVersion = 0
	}

	config.mu.Lock()
	if args.Exists("-migrate") {
		err := db.Debug().Unscoped().AutoMigrate(&SettingDomain{}, &Setting{}).Error
		if err != nil {
			dbVersion = 1
		}
	}

	switch dbVersion {
	case 0:
		config.data0 = make(map[string]map[string]generic.Value)
		var items []SettingLegacy
		db.Find(&items)
		for _, item := range items {
			item.Domain = strings.ToUpper(item.Domain)
			item.Name = strings.ToUpper(item.Name)
			if _, ok := config.data0[item.Domain]; !ok {
				config.data0[item.Domain] = map[string]generic.Value{}
			}
			config.data0[item.Domain][item.Name] = generic.Parse(item.Value)
		}
		var list []SettingDomain
		db.Find(&list)
		for idx, item := range list {
			domains[item.Domain] = list[idx]
		}
	case 1:
		config.data1 = make(map[string]generic.Value)
		var items []SettingDomain
		// This preload limits the depth of children domains to 2 level, domain, subdomain and parameter
		err := db.Model(&SettingDomain{}).Preload("Parameters").Preload("ChildrenDomains").Preload("ChildrenDomains.Parameters").Where("parent_domain = 0 OR parent_domain IS NULL").Find(&items).Error
		if err != nil {
			log.Error("Error while loading settings from db", err)
		}

		var parameters map[string]generic.Value
		for _, domain := range items {
			parameters = enumerateSettings(domain)
			for k, v := range parameters {
				config.data1[k] = v
			}
		}
		break
	}

	config.mu.Unlock()

	return nil
}

func enumerateSettings(dom SettingDomain) map[string]generic.Value {
	var toReturn = make(map[string]generic.Value)
	for _, param := range dom.Parameters {
		var toSet generic.Value
		if param.Value == "" {
			toSet = generic.Parse(param.DefaultValue)
		} else {
			toSet = generic.Parse(param.Value)
		}
		toReturn[dom.Domain+"."+param.Name] = toSet
	}

	for _, child := range dom.ChildrenDomains {
		toAdd := enumerateSettings(child)
		for key, value := range toAdd {
			toReturn[dom.Domain+"."+key] = value
		}
	}

	return toReturn
}
