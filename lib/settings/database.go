package settings

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/db"
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

// LoadDatabaseSettings loads settings from the database.
// Settings are organized by domains and loaded with their full hierarchical path.
// If the settings or settings_domain tables do not exist yet, the function returns
// silently without an error — the tables are optional and created during migration.
func LoadDatabaseSettings() error {
	migrator := db.GetInstance().Migrator()
	if !migrator.HasTable(&Setting{}) || !migrator.HasTable(&SettingDomain{}) {
		return nil
	}

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

	// Populate settings with full hierarchical paths
	for _, setting := range settings {
		setting.SettingsDomain = domainMap[setting.DomainID]
		domainPath := getFullDomainPath(setting.SettingsDomain)
		fullKey := domainPath + "." + setting.Name
		setData(fullKey, setting.Value)
	}

	return nil
}

// saveSingleSetting saves a single setting to the database.
// Creates or updates the setting in the default domain.
func saveSingleSetting(key string, value any) error {
	// Ensure default domain exists (cached after first call)
	var defaultDomain SettingDomain
	err := db.Where("domain = ?", "default").FirstOrCreate(&defaultDomain, SettingDomain{
		Domain:      "default",
		Title:       "Default Settings",
		Description: "Default settings domain",
		ReadOnly:    false,
		Visible:     true,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to create default domain: %w", err)
	}

	// Convert value to string for storage
	valueStr := fmt.Sprint(value)

	// Find or create the setting
	var setting Setting
	err = db.Where("domain_id = ? AND name = ?", defaultDomain.SettingsDomainID, key).
		FirstOrCreate(&setting, Setting{
			DomainID: defaultDomain.SettingsDomainID,
			Name:     key,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to create/find setting %s: %w", key, err)
	}

	// Update value if changed
	if setting.Value != valueStr {
		setting.Value = valueStr
		if err := db.Save(&setting).Error; err != nil {
			return fmt.Errorf("failed to save setting %s: %w", key, err)
		}
	}

	return nil
}

// saveDatabaseSettings saves multiple settings to the database.
// Creates or updates settings in the default domain (domain_id = 1).
func saveDatabaseSettings(flattenedData map[string]any) error {
	// Ensure default domain exists
	var defaultDomain SettingDomain
	err := db.Where("domain = ?", "default").FirstOrCreate(&defaultDomain, SettingDomain{
		Domain:      "default",
		Title:       "Default Settings",
		Description: "Default settings domain",
		ReadOnly:    false,
		Visible:     true,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to create default domain: %w", err)
	}

	// Save or update each setting
	for key, value := range flattenedData {
		// Convert value to string for storage
		valueStr := fmt.Sprint(value)

		var setting Setting
		err := db.Where("domain_id = ? AND name = ?", defaultDomain.SettingsDomainID, key).
			FirstOrCreate(&setting, Setting{
				DomainID: defaultDomain.SettingsDomainID,
				Name:     key,
			}).Error
		if err != nil {
			return fmt.Errorf("failed to create/find setting %s: %w", key, err)
		}

		// Update value if changed
		if setting.Value != valueStr {
			setting.Value = valueStr
			if err := db.Save(&setting).Error; err != nil {
				return fmt.Errorf("failed to save setting %s: %w", key, err)
			}
		}
	}

	return nil
}
