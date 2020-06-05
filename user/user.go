package user

import (
	"fmt"
	"github.com/iesreza/io/lib/log"
	"github.com/iesreza/validate"
	"github.com/jinzhu/gorm"
	"github.com/nbutton23/zxcvbn-go"
	"gopkg.in/hlandau/passlib.v1"
)

var db *gorm.DB
var config struct {
	App struct {
		StrongPass int
	}
}

// Save save user instance
func (u *User) Save() error {
	if u.Group == nil && u.GroupID > 0 {
		gp := Group{}
		if db.Where("id = ?", u.GroupID).Find(&gp).RecordNotFound() {
			return fmt.Errorf("invalid group id")
		}
		u.Group = &gp
	}

	err := validate.Validate(u)
	if err != nil {
		return err
	}

	temp := User{}
	if !db.Where("username = ?", u.Username).Find(&temp).RecordNotFound() {
		if u.ID == 0 || (u.ID > 0 && u.ID != temp.ID) {
			return fmt.Errorf("username exist")
		}
	}

	if !db.Where("email = ?", u.Email).Find(&temp).RecordNotFound() {
		if u.ID == 0 || (u.ID > 0 && u.ID != temp.ID) {
			return fmt.Errorf("email exist")
		}

	}
	if u.ID > 0 {
		return db.Save(&u).Error
	} else {
		score := zxcvbn.PasswordStrength(u.Password, nil).Score
		if score < config.App.StrongPass {
			return fmt.Errorf("password is not strength (%d)", score)
		}
		u.SetPassword(u.Password)
		return db.Create(&u).Error
	}
}

// AfterFind after find event
func (u *User) AfterFind() (err error) {
	gp := Group{}
	if !db.Where("id = ?", u.GroupID).Find(&gp).RecordNotFound() {
		u.Group = &gp
	}

	if u.Group != nil {
		u.Roles = u.Group.Roles
	}
	return
}

// SetPassword set user password
func (u *User) SetPassword(password string) error {
	pass, err := passlib.Hash(password)
	if err != nil {
		return err
	}
	u.Password = pass
	return nil
}

// HasRole check if user has role
func (u *User) HasRole(v interface{}) bool {
	if u.Admin {
		return true
	}
	if u.Anonymous {
		return false
	}
	var role Role
	if val, ok := v.(Role); ok {
		role = val
	} else {
		if db.Where("id = ? OR code_name = ?", val, val).Take(&role).RecordNotFound() {
			return false
		}
	}
	for _, rl := range u.Roles {
		if rl.CodeName == role.CodeName {
			return true
		}
	}
	return false
}

// HasPerm check if user has permission
func (u *User) HasPerm(v string) bool {
	if u.Admin {
		return true
	}
	if u.Anonymous {
		return false
	}
	return u.Group.HasPerm(v)
}

// Image return user image
func (u User) Image() string {
	return "files/profile/profile-" + fmt.Sprint(u.ID) + ".jpg"
}

// SetGroup set user group
func (u *User) SetGroup(group interface{}) error {
	var gp Group
	if val, ok := group.(Group); ok {
		gp = val
	} else {
		if db.Where("id = ? OR code_name = ?", group, group).Take(&gp).RecordNotFound() {
			return fmt.Errorf("group not exist")
		}
	}
	u.Group = &gp
	return db.Save(u).Error
}

// Sync synchronize permissions on startup
func (perms Permissions) Sync(app string) {
	var ids []uint
	for _, perm := range perms {
		perm.CodeName = app + "." + perm.CodeName
		perm.App = app
		err := validate.Validate(perm)
		if err != nil {
			log.Error("Permission validation error")
			log.Error(err)
			continue
		}
		p := Permission{}
		if db.Where("code_name = ?", perm.CodeName).Take(&p).RecordNotFound() {
			db.Create(&perm)
			ids = append(ids, perm.ID)
		} else {
			ids = append(ids, p.ID)
		}

	}

	var removed []Permission
	db.Where("app = ? AND id NOT IN (?)", app, ids).Find(&removed)
	for _, item := range removed {
		db.Delete(RolePermission{}, "permission_id = ?", item.ID)
		db.Delete(Permission{}, "id = ?", item.ID)
	}

	if len(removed) > 0 {
		updateRolePermissions()
	}
}
