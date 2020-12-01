package evo

import (
	"errors"
	"fmt"
	"github.com/getevo/evo/lib/log"
	"github.com/getevo/evo/lib/validate"
	"gorm.io/gorm"
)

var memoryRolePermissions = cMap{}

func updateRolePermissions() {
	memoryRolePermissions.Init()
	var roles []Role
	db.Find(&roles)
	for _, role := range roles {
		var rolePerms []RolePermission
		db.Where("role_id = ?", role.ID).Find(&rolePerms)
		var permissions Permissions
		for _, perm := range rolePerms {
			var p Permission
			if errors.Is(db.Where("id = ?", perm.PermissionID).Take(&p).Error, gorm.ErrRecordNotFound) {
				log.Warning("Roles: found inconsistency, automatically remove permission id %d to fix.", perm.PermissionID)
				db.Delete(RolePermission{}, "id = ?", perm.PermissionID)
			} else {
				permissions = append(permissions, p)
			}
		}
		memoryRolePermissions.Set(role.ID, &permissions)
	}
}

// AfterFind after find event
func (r *Role) AfterFind(tx *gorm.DB) (err error) {
	perms := memoryRolePermissions.Get(r.ID)
	if perms != nil {
		r.Permission = perms.(*Permissions)
		r.PermissionSet = []string{}
		for _, item := range *r.Permission {
			r.PermissionSet = append(r.PermissionSet, item.CodeName)
		}
		return
	}
	r.Permission = &Permissions{}
	r.PermissionSet = []string{}
	return
}

// Save save role instance
func (r *Role) Save() error {
	var perms Permissions
	for _, item := range r.PermissionSet {
		var perm Permission
		if errors.Is(db.Where("code_name = ?", item).Take(&perm).Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("permission not found")
		}
		perms = append(perms, perm)
	}
	err := validate.Validate(r)
	if err != nil {
		return err
	}
	temp := Role{}
	if !errors.Is(db.Where("code_name = ?", r.CodeName).Find(&temp).Error, gorm.ErrRecordNotFound) {
		if r.ID == 0 || (r.ID > 0 && r.ID != temp.ID) {
			return fmt.Errorf("codename exist")
		}
	}
	if r.ID > 0 {
		r.SetPermission(perms)
		return db.Save(&r).Error
	} else {
		err := db.Create(&r).Error
		if err != nil {
			return err
		}
		return r.SetPermission(perms)
	}
}

// HasPerm check if role has permission
func (r *Role) HasPerm(v string) bool {
	for _, item := range r.PermissionSet {
		if item == v {
			return true
		}
	}
	return false
}

// SetPermission set role permission
func (r *Role) SetPermission(permissions Permissions) error {

	var listId []uint
	for k, item := range permissions {
		var perm RolePermission
		if errors.Is(db.Where("role_id = ? AND permission_id = ?", r.ID, item.ID).Take(&perm).Error, gorm.ErrRecordNotFound) {
			err := db.Create(&RolePermission{RoleID: r.ID, PermissionID: item.ID}).Error
			if err != nil {
				return err
			}
		}
		listId = append(listId, permissions[k].ID)
	}

	memoryRolePermissions.Set(r.ID, &permissions)
	return db.Delete(RolePermission{}, "role_id = ? AND permission_id NOT IN (?)", r.ID, listId).Error
}
