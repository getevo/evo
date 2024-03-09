package database

import (
	"time"
)

type SettingDomain struct {
	ID              uint            `gorm:"column:domain_id;primarykey" json:"domain_id"`
	Domain          string          `gorm:"column:domain;type:VARCHAR(50);index:key,unique" json:"domain"`
	Title           string          `gorm:"column:title" json:"title"`
	Description     string          `gorm:"column:description" json:"description"`
	ReadOnly        bool            `gorm:"column:read_only" json:"read_only"`
	Visible         bool            `gorm:"column:visible" json:"visible"`
	ParentDomain    *uint           `gorm:"column:parent_domain;index:key,unique;default:NULL" json:"parent_domain"`
	ChildrenDomains []SettingDomain `gorm:"foreignkey:ParentDomain;references:ID" json:"children_domains"`
	Parameters      []Setting       `gorm:"foreignkey:DomainID;references:ID" json:"parameters"`
	CreatedAt       time.Time       `json:"-"`
	UpdatedAt       time.Time       `json:"-"`
}

func (SettingDomain) TableName() string {
	return "settings_domain"
}

type Setting struct {
	ID           uint      `gorm:"primarykey" json:"-"`
	DomainID     uint      `gorm:"column:domain_id;index:key,unique" json:"domain"`
	Value        string    `gorm:"column:value" json:"value"`
	DefaultValue string    `gorm:"column:default_value" json:"default_value"`
	Description  string    `gorm:"column:description" json:"description"`
	Name         string    `gorm:"column:name;index:key,unique" json:"name"`
	ReadOnly     bool      `gorm:"column:read_only" json:"read_only"`
	Title        string    `gorm:"column:title" json:"title"`
	Visible      bool      `gorm:"column:visible" json:"visible"`
	Type         string    `gorm:"column:type" json:"type"`
	Params       string    `gorm:"column:params" json:"params"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (Setting) TableName() string {
	return "settings"
}
