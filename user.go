package evo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getevo/evo/lib/validate"
	"gorm.io/gorm"
)

var db *gorm.DB

/*var config struct {
	App struct {
		StrongPass int
	}
}*/

// TODO: parse using reflect
// Save save user instance
func (u *User) ParseParams(out interface{}) error {
	b, err := json.Marshal(u.Params)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// Save save user instance
func (u *User) Save() error {
	return userInterface.Save(u)
}

// AfterFind after find event
func (u *User) AfterFind(tx *gorm.DB) (err error) {
	return userInterface.AfterFind(u)
}

// SetPassword set user password
func (u *User) SetPassword(password string) error {
	return userInterface.SetPassword(u, password)
}

// HasRole check if user has role
func (u *User) HasRole(v interface{}) bool {
	return userInterface.HasRole(u, v)
}

// HasPerm check if user has permission
func (u *User) HasPerm(v string) bool {
	return userInterface.HasPerm(u, v)
}

// Image return user image
func (u *User) Image() string {
	return userInterface.Image(u)
}

// SetGroup set user group
func (u *User) SetGroup(group interface{}) error {
	return userInterface.SetGroup(u, group)
}

// Sync synchronize permissions on startup
func (perms Permissions) Sync(app string) {
	userInterface.SyncPermissions(app, perms)
}

// SetGroup set user group
func (u *User) FromRequest(r *Request) {
	userInterface.FromRequest(r)
}

// Save save the group instance
func (g *UserGroup) Save() error {
	var set []*Role
	var remove []*Role
	for _, rl := range g.RoleSet {
		item := Role{}
		if !errors.Is(db.Where("id = ? OR code_name = ?", rl, rl).Take(&item).Error, gorm.ErrRecordNotFound) {
			set = append(set, &item)
		}
	}

	if g.ID > 0 {
		for _, old := range g.Roles {
			found := false
			for _, set := range g.Roles {
				if set.ID == old.ID {
					found = true
					break
				}
			}
			if !found {
				remove = append(remove, old)
			}
		}

	}

	g.Roles = set
	err := validate.Validate(g)
	if err != nil {
		return err
	}
	temp := UserGroup{}
	if !errors.Is(db.Where("code_name = ?", g.CodeName).Take(&temp).Error, gorm.ErrRecordNotFound) {
		if g.ID == 0 || (g.ID > 0 && g.ID != temp.ID) {
			return fmt.Errorf("codename exist")
		}
	}
	if g.ID > 0 {
		if len(remove) > 0 {
			err = db.Model(g).Association("Roles").Delete(remove)
			if err != nil {
				return err
			}
		}
		return db.Save(g).Error
	} else {
		return db.Create(g).Error
	}
}

// HasPerm check the group if has permission
func (g *UserGroup) HasPerm(v string) bool {
	for _, role := range g.Roles {
		if role.HasPerm(v) {
			return true
		}
	}
	return false
}

// AfterFind after find event
func (g *UserGroup) AfterFind(tx *gorm.DB) (err error) {
	var roles []*Role
	//db.Model(g).Related(&roles, "Roles")
	g.Roles = roles
	if len(g.Roles) > 0 {
		for _, item := range g.Roles {
			g.RoleSet = append(g.RoleSet, item.CodeName)
		}
	}
	return
}
