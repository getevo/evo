package evo

import (
	"errors"
	"fmt"
	"github.com/getevo/evo/lib/jwt"
	"github.com/getevo/evo/lib/log"
	"github.com/getevo/evo/lib/validate"
	"github.com/nbutton23/zxcvbn-go"
	"gopkg.in/hlandau/passlib.v1"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type LocalUserImplementation struct{}
type UserInterface interface {
	Save(u *User) error
	HasPerm(u *User, v string) bool
	HasRole(u *User, v interface{}) bool
	Image(u *User) string
	SetPassword(u *User, password string) error
	SetGroup(u *User, group interface{}) error
	AfterFind(u *User) error
	SyncPermissions(app string, permissions Permissions)
	FromRequest(r *Request)
}

var userInterface UserInterface = LocalUserImplementation{}

func SetUserInterface(p UserInterface) {
	userInterface = p
}

func (p LocalUserImplementation) Save(u *User) error {
	if u.Group == nil && u.GroupID > 0 {
		gp := UserGroup{}
		if errors.Is(db.Where("id = ?", u.GroupID).Find(&gp).Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("invalid group id")
		}
		u.Group = &gp
	}

	err := validate.Validate(u)
	if err != nil {
		return err
	}

	temp := User{}
	if !errors.Is(db.Where("username = ?", u.Username).Find(&temp).Error, gorm.ErrRecordNotFound) {
		if u.ID == 0 || (u.ID > 0 && u.ID != temp.ID) {
			return fmt.Errorf("username exist")
		}
	}

	if !errors.Is(db.Where("email = ?", u.Email).Find(&temp).Error, gorm.ErrRecordNotFound) {
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
func (p LocalUserImplementation) HasPerm(u *User, v string) bool {
	if u.Admin {
		return true
	}
	if u.Anonymous {
		return false
	}
	return u.Group.HasPerm(v)
}
func (p LocalUserImplementation) HasRole(u *User, v interface{}) bool {
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
		if errors.Is(db.Where("id = ? OR code_name = ?", val, val).Take(&role).Error, gorm.ErrRecordNotFound) {
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
func (p LocalUserImplementation) Image(u *User) string {
	return "files/profile/profile-" + fmt.Sprint(u.ID) + ".jpg"
}
func (p LocalUserImplementation) SetPassword(u *User, password string) error {
	pass, err := passlib.Hash(password)
	if err != nil {
		return err
	}
	u.Password = pass
	return nil
}
func (p LocalUserImplementation) SetGroup(u *User, group interface{}) error {
	var gp UserGroup
	if val, ok := group.(UserGroup); ok {
		gp = val
	} else {
		if errors.Is(db.Where("id = ? OR code_name = ?", group, group).Take(&gp).Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("group not exist")
		}
	}
	u.Group = &gp
	return db.Save(u).Error

}
func (p LocalUserImplementation) AfterFind(u *User) error {
	gp := UserGroup{}
	if !errors.Is(db.Where("id = ?", u.GroupID).Find(&gp).Error, gorm.ErrRecordNotFound) {
		u.Group = &gp
	}

	if u.Group != nil {
		u.Roles = u.Group.Roles
	}
	return nil
}
func (p LocalUserImplementation) SyncPermissions(app string, perms Permissions) {
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
		if errors.Is(db.Where("code_name = ?", perm.CodeName).Take(&p).Error, gorm.ErrRecordNotFound) {
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

// SetGroup set user group
func (u LocalUserImplementation) FromRequest(r *Request) {
	r.User = &User{Anonymous: true}
	token := ""
	token = r.Cookies("access_token")
	if strings.TrimSpace(token) == "" {
		token = r.Cookies("auth_token")
	}
	if token == "" {
		token = r.Get("Authorization")
	}
	if token == "" {
		token = r.Cookies("Authorization")
	}
	if token != "" {
		token, err := jwt.Verify(token)
		if err == nil {
			r.JWT = &token
			r.User = getUserFromJWT(&token)
		} else {
			r.SetCookie("access_token", "")
			r.Status(http.StatusUnauthorized)
			r.Send("invalid JWT token")
			log.Error(err)
		}
	} else {
		r.JWT = &jwt.Payload{Empty: true, Data: map[string]interface{}{}}
	}

}

func getUserFromJWT(payload *jwt.Payload) *User {
	var user User
	// return user using jwt
	if payload.Data != nil {
		if id, ok := payload.Data["id"]; ok {
			Database.Where("id = ?", id).Take(&user)
			if user.ID == 0 {
				user.Anonymous = true
			}
		}
	} else {
		user.Anonymous = false
	}

	return &user
}
