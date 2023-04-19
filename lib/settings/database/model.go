package database

type SettingDomain struct {
	Domain      string    `gorm:"column:domain;primaryKey;size:255" json:"domain"`
	Title       string    `gorm:"column:title" json:"title"`
	Description string    `gorm:"column:description" json:"description"`
	ReadOnly    bool      `gorm:"column:read_only" json:"read_only"`
	Visible     bool      `gorm:"column:visible" json:"visible"`
	Items       []Setting `gorm:"-"`
}

func (SettingDomain) TableName() string {
	return "settings_domain"
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

func (Setting) TableName() string {
	return "settings"
}
