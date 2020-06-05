package user

import (
	"encoding/json"
	"github.com/iesreza/io/lib/data"
	"github.com/iesreza/io/lib/text"
	"github.com/jinzhu/gorm"
	"time"
)

// Model common model stuff
type Model struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}

// Role role struct
type Role struct {
	Model
	Name          string       `json:"name" form:"name" validate:"empty=false & format=strict_html"`
	CodeName      string       `json:"code_name" json:"code_name" validate:"empty=false & format=slug" gorm:"type:varchar(100);unique_index"`
	Parent        uint         `json:"parent" form:"parent"`
	Groups        []*Group     `json:"-" gorm:"many2many:group_roles;"`
	Permission    *Permissions `json:"permissions_data" gorm:"-"  validate:"-"`
	PermissionSet []string     `json:"permissions" form:"permissions" gorm:"-"`
}

// Group group struct
type Group struct {
	Model
	Name     string   `json:"name" form:"name" validate:"empty=false & format=strict_html"`
	CodeName string   `json:"code_name" json:"code_name" validate:"empty=false & format=slug" gorm:"type:varchar(100);unique_index"`
	Parent   uint     `json:"parent" form:"parent"`
	Roles    []*Role  `json:"roles_data" gorm:"many2many:group_roles;" validate:"-"`
	RoleSet  []string `json:"roles" form:"roles" gorm:"-"`
}

// Permission permission struct
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
type RolePermission struct {
	Model
	RoleID       uint
	PermissionID uint
}

// User user struct
type User struct {
	Model
	Name      string    `json:"name" form:"name"`
	Username  string    `json:"username" form:"username"  validate:"empty=false | format=username" gorm:"type:varchar(32);unique_index"`
	Password  string    `json:"-" form:"-" validate:"empty=false & format=strict_html"`
	Email     string    `json:"email" form:"email" validate:"empty=false & format=email" gorm:"type:varchar(32);unique_index"`
	Roles     []*Role   `json:"roles" form:"roles"  validate:"-"`
	Group     *Group    `json:"-" form:"-" gorm:"-"  validate:"-"`
	GroupID   uint      `json:"group_id" form:"group_id"`
	Anonymous bool      `json:"anonymous" form:"anonymous" gorm:"-"`
	Active    bool      `json:"active" form:"active"`
	Seen      time.Time `json:"seen" form:"seen"`
	Admin     bool      `json:"admin" form:"admin"`
	Params    data.Map  `gorm:"type:json" form:"params" json:"params"`
}

// TableName return role model table name
func (Role) TableName() string {
	return "role"
}

// TableName return group model table name
func (Group) TableName() string {
	return "group"
}

// InitUserModel initialize the user model with given config
func InitUserModel(database *gorm.DB, config interface{}) {
	j := text.ToJSON(config)
	json.Unmarshal([]byte(j), &config)
	db = database
	db.AutoMigrate(&User{}, &Group{}, &Role{}, &Permission{}, &RolePermission{})
	updateRolePermissions()

}
