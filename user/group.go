package user

import (
	"fmt"
	"github.com/iesreza/validate"
)

// Save save the group instance
func (g *Group) Save() error {
	var set []*Role
	var remove []*Role
	for _, rl := range g.RoleSet {
		item := Role{}
		if !db.Where("id = ? OR code_name = ?", rl, rl).Take(&item).RecordNotFound() {
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
	temp := Group{}
	if !db.Where("code_name = ?", g.CodeName).Take(&temp).RecordNotFound() {
		if g.ID == 0 || (g.ID > 0 && g.ID != temp.ID) {
			return fmt.Errorf("codename exist")
		}
	}
	if g.ID > 0 {
		if len(remove) > 0 {
			err = db.Model(g).Association("Roles").Delete(remove).Error
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
func (g *Group) HasPerm(v string) bool {
	for _, role := range g.Roles {
		if role.HasPerm(v) {
			return true
		}
	}
	return false
}

// AfterFind after find event
func (g *Group) AfterFind() (err error) {
	var roles []*Role
	db.Model(g).Related(&roles, "Roles")
	g.Roles = roles
	if len(g.Roles) > 0 {
		for _, item := range g.Roles {
			g.RoleSet = append(g.RoleSet, item.CodeName)
		}
	}
	return
}
