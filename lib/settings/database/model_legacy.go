package database

type SettingDomainLegacy struct {
	Domain      string    `gorm:"column:domain;primaryKey;size:255" json:"domain"`
	Title       string    `gorm:"column:title;size:255" json:"title"`
	Description string    `gorm:"column:description;size:255" json:"description"`
	ReadOnly    bool      `gorm:"column:read_only" json:"read_only"`
	Visible     bool      `gorm:"column:visible" json:"visible"`
	Items       []Setting `gorm:"-"`
}

func (SettingDomainLegacy) TableName() string {
	return "settings_domain"
}

type SettingLegacy struct {
	Domain      string `gorm:"column:domain;primaryKey;size:255" json:"domain"`
	Name        string `gorm:"column:name;primaryKey;size:255" json:"name"`
	Title       string `gorm:"column:title;size:255" json:"title"`
	Description string `gorm:"column:description;size:255" json:"description"`
	Value       string `gorm:"column:value" json:"value"`
	Type        string `gorm:"column:type;size:255" json:"type"`
	Params      string `gorm:"column:params;size:512" json:"params"`
	ReadOnly    bool   `gorm:"column:read_only" json:"read_only"`
	Visible     bool   `gorm:"column:visible" json:"visible"`
}

func (SettingLegacy) TableName() string {
	return "settings"
}
