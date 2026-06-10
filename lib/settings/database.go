package settings

import (
	"fmt"
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
	// The primary key column is `id`. Without the explicit column tag GORM
	// derives `settings_id`, which does not exist in the table — every
	// INSERT/UPDATE then references a missing column and fails (reads via
	// SELECT * still work, which is why this stayed hidden). Keep this in sync
	// with the canonical model in apps/settings/models.go.
	SettingsID     uint          `gorm:"column:id;primaryKey" json:"-"`
	DomainID       uint          `gorm:"column:domain_id;fk:settings_domain" json:"domain"`
	Domain         string        `gorm:"-" json:"-"`
	Name           string        `gorm:"column:name;size:128" json:"name"`
	Title          string        `gorm:"column:title" json:"title"`
	Description    string        `gorm:"column:description" json:"description"`
	Value          string        `gorm:"column:value" json:"value"`
	DefaultValue   string        `gorm:"column:default_value" json:"default_value"`
	ReadOnly       bool          `gorm:"column:read_only" json:"read_only"`
	Visible        bool          `gorm:"column:visible" json:"visible"`
	// No `protected` column exists in the table; mapping it to a column made
	// full-struct writes fail. Kept as a non-persisted field for API/JSON.
	Protected      bool          `gorm:"-" json:"protected"`
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

// splitSettingKey splits a hierarchical setting key into its domain path and
// leaf name on the LAST dot separator. This mirrors LoadDatabaseSettings, which
// reconstructs a key as "<domainPath>.<name>". A key with no dot has an empty
// domain path and is stored in the default domain.
//
//	"CMS.LLMS" -> ("CMS", "LLMS")
//	"A.B.C"    -> ("A.B", "C")
//	"FLATKEY"  -> ("",    "FLATKEY")
func splitSettingKey(key string) (domainPath, name string) {
	if idx := strings.LastIndex(key, "."); idx >= 0 {
		return key[:idx], key[idx+1:]
	}
	return "", key
}

// resolveDomainID resolves a (possibly nested) domain path to its domain id,
// creating any missing domains along the way. An empty path resolves to the
// "default" domain. This is the write-side counterpart of getFullDomainPath
// used by LoadDatabaseSettings, so a value written by Set/SetMulti is read back
// under the exact same key after a reload (rather than landing in an orphan
// row that nothing reads — the previous behaviour, which flattened the whole
// dotted key into a single name under the "default" domain).
func resolveDomainID(domainPath string) (uint, error) {
	if domainPath == "" {
		// Search only on the domain name; create-only columns go in Attrs so a
		// pre-existing "default" row matches regardless of its other column
		// values (see the loop below for why this matters).
		var defaultDomain SettingDomain
		err := db.Where("domain = ?", "default").
			Attrs(SettingDomain{
				Title:       "Default Settings",
				Description: "Default settings domain",
				Visible:     true,
			}).
			FirstOrCreate(&defaultDomain).Error
		if err != nil {
			return 0, fmt.Errorf("failed to create default domain: %w", err)
		}
		return defaultDomain.SettingsDomainID, nil
	}

	var parentID *uint
	var leafID uint
	for _, segment := range strings.Split(domainPath, ".") {
		// Search ONLY on the domain's identity (name + parent). Create-only
		// columns go in Attrs — if they were passed as FirstOrCreate conds GORM
		// would fold their non-zero values (title, visible) into the SELECT, so
		// an existing domain seeded by a migration with a different title would
		// fail to match and a duplicate domain would be created.
		var domain SettingDomain
		query := db.Where("domain = ?", segment)
		if parentID == nil {
			query = query.Where("parent_domain IS NULL")
		} else {
			query = query.Where("parent_domain = ?", *parentID)
		}
		err := query.
			Attrs(SettingDomain{
				ParentDomain: parentID,
				Title:        segment,
				Visible:      true,
			}).
			FirstOrCreate(&domain).Error
		if err != nil {
			return 0, fmt.Errorf("failed to create/find domain %q: %w", segment, err)
		}
		leafID = domain.SettingsDomainID
		id := domain.SettingsDomainID
		parentID = &id
	}
	return leafID, nil
}

// saveSingleSetting saves a single setting to the database under the domain
// implied by its hierarchical key. Creates the domain hierarchy and the
// setting row if they don't exist; otherwise updates the existing row in place
// so the change is picked up by every node via the settings-table CDC hook.
func saveSingleSetting(key string, value any) error {
	domainPath, name := splitSettingKey(key)
	domainID, err := resolveDomainID(domainPath)
	if err != nil {
		return err
	}

	// Convert value to string for storage
	valueStr := fmt.Sprint(value)

	// Find or create the setting under the resolved domain
	var setting Setting
	err = db.Where("domain_id = ? AND name = ?", domainID, name).
		FirstOrCreate(&setting, Setting{
			DomainID: domainID,
			Name:     name,
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

// saveDatabaseSettings saves multiple settings to the database, each under the
// domain implied by its hierarchical key (see saveSingleSetting).
func saveDatabaseSettings(data map[string]any) error {
	for key, value := range data {
		if err := saveSingleSetting(key, value); err != nil {
			return err
		}
	}
	return nil
}
