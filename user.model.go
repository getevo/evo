package evo

import (
	"encoding/json"
	"github.com/getevo/evo/lib/data"
	"github.com/getevo/evo/lib/text"
	"gorm.io/gorm"
	"time"
)

/*// Model common model stuff
type Model struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}*/

// Role role struct
// @doc	type 	model
type Role struct {
	Model
	Name          string       `json:"name" form:"name" validate:"empty=false & format=strict_html"`
	CodeName      string       `json:"code_name" json:"code_name" validate:"empty=false & format=slug" gorm:"type:varchar(100);unique_index"`
	Parent        uint         `json:"parent" form:"parent"`
	Groups        []*UserGroup `json:"-" gorm:"many2many:group_roles;"`
	Permission    *Permissions `json:"permissions_data" gorm:"-"  validate:"-"`
	PermissionSet []string     `json:"permissions" form:"permissions" gorm:"-"`
}

// Group group struct
// @doc type 			model
type UserGroup struct {
	Model
	Name     string   `json:"name" form:"name" validate:"empty=false & format=strict_html"`
	CodeName string   `json:"code_name" json:"code_name" validate:"empty=false & format=slug" gorm:"type:varchar(100);unique_index"`
	Parent   uint     `json:"parent" form:"parent"`
	Roles    []*Role  `json:"roles_data" gorm:"many2many:group_roles;" validate:"-"`
	RoleSet  []string `json:"roles" form:"roles" gorm:"-"`
}

// Permission permission struct
// @doc type 			model
type Permission struct {
	Model
	CodeName    string `json:"code_name" form:"code_name" validate:"empty=false"`
	Title       string `json:"title" form:"title" validate:"empty=false"`
	Description string `json:"description" form:"description"`
	App         string `json:"app" form:"app" validate:"empty=false"`
}

// Permissions slice of permissions
type Permissions []Permission

// RolePermission role to permission orm
// @doc type 			model
type RolePermission struct {
	Model
	RoleID       uint
	PermissionID uint
}

// User user struct
// @doc type 			model
type User struct {
	Model
	GivenName  string     `json:"given_name" form:"given_name" mapstructure:"given_name"`
	FamilyName string     `json:"family_name" form:"family_name" mapstructure:"family_name"`
	Username   string     `json:"username" form:"username"  mapstructure:"preferred_username"  validate:"empty=false | format=username" gorm:"type:varchar(32);unique_index"`
	Password   string     `json:"-" form:"-" validate:"empty=false & format=strict_html"`
	Email      string     `json:"email" form:"email" mapstructure:"email" validate:"empty=false & format=email" gorm:"type:varchar(32);unique_index"`
	Roles      []*Role    `json:"roles" form:"roles"  validate:"-"`
	Group      *UserGroup `json:"-" form:"-" gorm:"-"  validate:"-"`
	GroupID    uint       `json:"group_id" form:"group_id"`
	Anonymous  bool       `json:"anonymous" form:"anonymous" gorm:"-"`
	Active     bool       `json:"active" form:"active"`
	Seen       time.Time  `json:"seen" form:"seen"`
	Admin      bool       `json:"admin" form:"admin"`
	Params     data.Map   `gorm:"type:json" mapstructure:"-" form:"params" json:"params"`
}

// TableName return role model table name
func (Role) TableName() string {
	return "role"
}

// TableName return group model table name
func (UserGroup) TableName() string {
	return "group"
}

// InitUserModel initialize the user model with given config
func InitUserModel(database *gorm.DB, config interface{}) {
	j := text.ToJSON(config)
	json.Unmarshal([]byte(j), &config)
	db = database

	if Arg.Migrate {
		db.AutoMigrate(&User{}, &UserGroup{}, &Role{}, &Permission{}, &RolePermission{})
	}
	updateRolePermissions()

}
