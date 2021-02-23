package auth

import (
	"fmt"
	"github.com/getevo/evo"
	e "github.com/getevo/evo/errors"
	"github.com/getevo/evo/lib/T"
	"github.com/getevo/evo/lib/constant"
	"github.com/getevo/evo/lib/jwt"
	"github.com/getevo/evo/lib/validate"
	"gopkg.in/hlandau/passlib.v1"
	"net/http"
	"time"
)

type Controller struct{}

// @doc type 			model
// @doc description		input parameters for login api
type AuthParams struct {
	Username string `json:"username" form:"username" validate:"empty=false"`
	Password string `json:"password" form:"password" validate:"empty=false"`
	Remember bool   `json:"remember" form:"remember"`
	Return   string `json:"return" form:"return" validate:"empty=true | one_of=json,text,html"`
	Redirect string `json:"redirect" form:"redirect"`
}

func GetUserByID(id interface{}) *evo.User {
	user := evo.User{}
	if db.Where("id = ?", id).Find(&user).RowsAffected == 0 {
		user.Anonymous = true
		return &user
	}

	return &user
}

func GetUserByUsername(username interface{}) *evo.User {
	user := evo.User{}
	if db.Where("username = ?", username).Find(&user).RowsAffected == 0 {
		return nil
	}
	return &user
}

func GetUserByEmail(email interface{}) *evo.User {
	user := evo.User{}
	if db.Where("email = ?", email).Find(&user).RowsAffected == 0 {
		return nil
	}
	return &user
}

func GetGroup(v interface{}) *evo.UserGroup {
	group := evo.UserGroup{}
	if db.Where("id = ? OR code_name", v, v).Find(&group).RowsAffected == 0 {
		return nil
	}
	return &group
}

func AuthUserByPassword(username, password string) (*evo.User, error) {
	user := GetUserByUsername(username)
	if user == nil {
		return user, fmt.Errorf("username not found")
	}
	_, err := passlib.Verify(password, user.Password)
	if err != nil {
		return user, fmt.Errorf("password not match")
	}
	user.Seen = time.Now()
	user.Save()
	return user, nil
}

func (c Controller) Login(r *evo.Request) {
	var err error
	var user *evo.User
	var token string

	r.Accepts("text/html", "application/json")
	params := AuthParams{}
	err = r.BodyParser(&params)
	if err != nil {
		r.WriteResponse(constant.ERROR_FORM_PARSE)
		return
	}
	if params.Return == "" {
		params.Return = "text"
	}

	err = validate.Validate(&params)
	if err == nil {
		user, err = AuthUserByPassword(params.Username, params.Password)
		if err == nil {
			extend := evo.GetConfig().JWT.Age
			if params.Remember {
				extend = 365 * 24 * time.Hour
			}
			token, err = jwt.Generate(map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"name":     user.GivenName,
				"seen":     user.Seen,
				"active":   user.Active,
				//"params":   user.Params,
			}, extend)
		}

	}
	switch params.Return {
	case "html":

		break
	case "json":
		if err != nil {

			r.WriteResponse(e.Field("form", "provided credentials is not valid"), err)
		} else {
			r.SetCookie("access_token", token)
			r.WriteResponse(true, "", token)
		}
		break
	case "text":
		if err != nil {
			r.Status(http.StatusBadRequest)
			r.Write("nok ")
			r.Write(err)
		} else {
			r.Write("ok")
			r.Write(token)
		}
		break
	}

}

func (c Controller) CreateUser(r *evo.Request) {

	if !r.User.HasPerm("auth.create.user") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var user = evo.User{}
	err := r.BodyParser(&user)

	if err == nil {
		err := user.Save()
		if err == nil {
			r.WriteResponse(map[string]interface{}{
				"id":          user.ID,
				"username":    user.Username,
				"given_name":  user.GivenName,
				"family_name": user.FamilyName,
				"param":       user.Params,
			},
			)
		} else {
			r.WriteResponse(e.Context(err))
		}
	} else {
		r.WriteResponse(constant.ERROR_FORM_PARSE)
	}

}

func (c Controller) CreateRole(r *evo.Request) {

	var role = evo.Role{}
	if !r.User.HasPerm("auth.create.role") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	err := r.BodyParser(&role)
	if err != nil {
		r.WriteResponse(constant.ERROR_FORM_PARSE)
		return
	}
	if err == nil {
		err := role.Save()
		if err == nil {
			r.WriteResponse(map[string]interface{}{
				"id":        role.ID,
				"name":      role.Name,
				"code_name": role.CodeName,
				"parent":    role.Parent,
			})
		} else {
			r.WriteResponse(err)
		}
	}

}

func (c Controller) CreateGroup(r *evo.Request) {

	var group = evo.UserGroup{}
	if !r.User.HasPerm("auth.create.group") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	err := r.BodyParser(&group)
	if err != nil {
		r.WriteResponse(constant.ERROR_FORM_PARSE)
		return
	}
	if err == nil {
		err := group.Save()
		if err == nil {
			r.WriteResponse(group)
		} else {
			r.WriteResponse(err)
		}
	}

}

// TODO: Password check and change
func (c Controller) EditUser(r *evo.Request) {

	if !r.User.HasPerm("auth.edit.user") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var user evo.User
	var id = T.Must(r.Params("id")).UInt()
	if db.Where("id = ?", id).Find(&user).RowsAffected == 0 {
		r.WriteResponse(constant.ERROR_INVALID_ID)
		return
	}
	err := r.BodyParser(&user)
	if err != nil {
		r.WriteResponse(constant.ERROR_FORM_PARSE)
		return
	}
	user.ID = id
	if err == nil {
		err := user.Save()
		if err == nil {
			r.WriteResponse(map[string]interface{}{
				"id":          user.ID,
				"username":    user.Username,
				"given_name":  user.GivenName,
				"family_name": user.FamilyName,
				"param":       user.Params,
			})
		} else {
			r.WriteResponse(err)
		}
	}

}

func (c Controller) EditRole(r *evo.Request) {

	if !r.User.HasPerm("auth.edit.role") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var role evo.Role
	var id = r.Params("id")

	if db.Where("id = ? OR code_name = ?", id, id).Find(&role).RowsAffected == 0 {
		r.WriteResponse(constant.ERROR_INVALID_ID)
		return
	}
	rid := role.ID
	err := r.BodyParser(&role)
	if err != nil {
		r.WriteResponse(constant.ERROR_FORM_PARSE)
		return
	}
	role.ID = rid
	if err == nil {
		err := role.Save()
		if err == nil {
			r.WriteResponse(role)
		} else {
			r.WriteResponse(err)
		}
	}

}

func (c Controller) EditGroup(r *evo.Request) {

	if !r.User.HasPerm("auth.edit.group") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var group evo.UserGroup
	var id = r.Params("id")
	if db.Where("id = ? OR code_name = ?", id, id).Find(&group).RowsAffected == 0 {
		r.WriteResponse(e.Field("id", constant.ERROR_INVALID_ID))
		return
	}

	var gid = group.ID
	err := r.BodyParser(&group)
	if err != nil {
		r.WriteResponse(constant.ERROR_FORM_PARSE)
		return
	}
	group.ID = gid
	if err == nil {
		err := group.Save()
		if err == nil {
			r.WriteResponse(group)
		} else {
			r.WriteResponse(err)
		}
	}

}

func (c Controller) GetGroups(r *evo.Request) {

	if !r.User.HasPerm("auth.group.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var groups []evo.UserGroup
	err := db.Find(&groups).Error
	if err != nil {
		r.WriteResponse(e.Field("id", constant.ERROR_INVALID_ID))
	} else {
		r.WriteResponse(groups)
	}
}

func (c Controller) GetGroup(r *evo.Request) {

	if !r.User.HasPerm("auth.group.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var group evo.UserGroup
	var id = r.Params("id")
	if db.Where("id = ? OR code_name = ?", id, id).Find(&group).RowsAffected == 0 {
		r.WriteResponse(e.Field("id", constant.ERROR_INVALID_ID))
		return
	} else {
		r.WriteResponse(group)
	}
}

func (c Controller) GetRoles(r *evo.Request) {

	if !r.User.HasPerm("auth.role.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var roles []evo.Role
	err := db.Find(&roles).Error
	if err != nil {
		r.WriteResponse(err)
	} else {
		r.WriteResponse(roles)
	}
}

func (c Controller) GetRole(r *evo.Request) {

	if !r.User.HasPerm("auth.role.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var role evo.Role
	var id = r.Params("id")
	if db.Where("id = ? OR code_name = ?", id, id).Find(&role).RowsAffected == 0 {
		r.WriteResponse(e.Field("id", constant.ERROR_INVALID_ID))
		return
	} else {
		r.WriteResponse(role)
	}
}

func (c Controller) GetRoleGroups(r *evo.Request) {

	if !r.User.HasPerm("auth.role.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var role evo.Role
	var id = r.Params("id")
	if db.Where("id = ? OR code_name = ?", id, id).Find(&role).RowsAffected == 0 {
		r.WriteResponse(e.Field("id", constant.ERROR_INVALID_ID))
		return
	} else {
		var groups []evo.UserGroup
		db.Joins(`INNER JOIN group_roles ON "group".id = group_roles.group_id`).Where("group_roles.role_id = ?", role.ID).Find(&groups)
		r.WriteResponse(groups)
	}
}

func (c Controller) GetUser(r *evo.Request) {

	if !r.User.HasPerm("auth.user.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var user evo.User
	var id = r.Params("id")
	if db.Where("id = ? OR username = ? OR email = ?", id, id, id).Find(&user).RowsAffected == 0 {
		r.WriteResponse(constant.ERROR_OBJECT_NOT_EXIST)
		return
	} else {
		r.WriteResponse(user)
	}
}

func (c Controller) GetMe(r *evo.Request) {

	r.WriteResponse(r.User)
}

func (c Controller) GetAllUsers(r *evo.Request) {

	if !r.User.HasPerm("auth.user.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var users []evo.User
	err := db.Offset(r.QueryI("offset").Int()).Limit(r.QueryI("limit").Int()).Find(&users).Error
	if err != nil {
		r.WriteResponse(err)
	} else {
		r.WriteResponse(users)
	}
}

func (c Controller) GetAllPermissions(r *evo.Request) {

	if !r.User.HasPerm("auth.role.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var perms []evo.Permission
	err := db.Find(&perms).Error
	if err != nil {
		r.WriteResponse(e.Context(err))
	} else {
		r.WriteResponse(perms)
	}
}
