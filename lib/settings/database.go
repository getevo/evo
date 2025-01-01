package settings

import (
	"github.com/getevo/evo/v2/lib/db"
	"strings"
	"time"
)

type SettingDomain struct {
	SettingsDomainID uint      `gorm:"column:domain_id;primarykey" json:"domain_id"`
	Domain           string    `gorm:"column:domain;type:VARCHAR(50)" json:"domain"`
	ParentDomain     *uint     `gorm:"column:parent_domain;fk:settings_domain" json:"parent_domain"`
	Title            string    `gorm:"column:title" json:"title"`
	Description      string    `gorm:"column:description" json:"description"`
	ReadOnly         bool      `gorm:"column:read_only" json:"read_only"`
	Visible          bool      `gorm:"column:visible" json:"visible"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

func (SettingDomain) TableName() string {
	return "settings_domain"
}

type Setting struct {
	SettingsID     uint          `gorm:"primaryKey" json:"-"`
	DomainID       uint          `gorm:"column:domain_id;fk:settings_domain" json:"domain"`
	Domain         string        `gorm:"-" json:"-"`
	Name           string        `gorm:"column:name;size:128" json:"name"`
	Title          string        `gorm:"column:title" json:"title"`
	Description    string        `gorm:"column:description" json:"description"`
	Value          string        `gorm:"column:value" json:"value"`
	DefaultValue   string        `gorm:"column:default_value" json:"default_value"`
	ReadOnly       bool          `gorm:"column:read_only" json:"read_only"`
	Visible        bool          `gorm:"column:visible" json:"visible"`
	Protected      bool          `gorm:"column:protected" json:"protected"`
	Type           string        `gorm:"column:type" json:"type"`
	Params         string        `gorm:"column:params" json:"params"`
	SettingsDomain SettingDomain `gorm:"-" json:"-"`
	CreatedAt      time.Time     `json:"-"`
	UpdatedAt      time.Time     `json:"-"`
}

func (Setting) TableName() string {
	return "settings"
}

func InitDatabaseSettings() {
	db.UseModel(Setting{}, SettingDomain{})
}

func LoadDatabaseSettings() error {
	var settings []Setting
	var domains []SettingDomain

	// Fetch all settings and domains from the database
	if err := db.Find(&settings).Error; err != nil {
		return err
	}
	if err := db.Find(&domains).Error; err != nil {
		return err
	}

	domainMap := make(map[uint]SettingDomain)
	for _, domain := range domains {
		domainMap[domain.SettingsDomainID] = domain
	}

	getFullDomainPath := func(domain SettingDomain) string {
		path := domain.Domain
		parentDomain := domain.ParentDomain

		for parentDomain != nil {
			parent, exists := domainMap[*parentDomain]
			if !exists {
				break
			}
			path = parent.Domain + "." + path
			parentDomain = parent.ParentDomain
		}
		return path
	}

	// Populate the nested map with settings
	for _, setting := range settings {
		setting.SettingsDomain = domainMap[setting.DomainID]
		domainPath := getFullDomainPath(setting.SettingsDomain)
		fullKey := domainPath + "." + setting.Name
		data[strings.ToUpper(fullKey)] = setting.Value
	}

	return nil
}
