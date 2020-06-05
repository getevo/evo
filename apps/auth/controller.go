package auth

import (
	"fmt"
	"github.com/gofiber/fiber"
	"github.com/iesreza/io"
	e "github.com/iesreza/io/errors"
	"github.com/iesreza/io/lib/T"
	"github.com/iesreza/io/lib/constant"
	"github.com/iesreza/io/lib/jwt"
	"github.com/iesreza/io/user"
	"github.com/iesreza/validate"
	"gopkg.in/hlandau/passlib.v1"
	"net/http"
	"time"
)

type Controller struct{}
type AuthParams struct {
	Username string `json:"username" form:"username" validate:"empty=false"`
	Password string `json:"password" form:"password" validate:"empty=false"`
	Remember bool   `json:"remember" form:"remember"`
	Return   string `json:"return" form:"return" validate:"empty=true | one_of=json,text,html"`
	Redirect string `json:"redirect" form:"redirect"`
}

func GetUserByID(id interface{}) *user.User {
	user := user.User{}
	if db.Where("id = ?", id).Find(&user).RecordNotFound() {
		user.Anonymous = true
		return &user
	}

	return &user
}

func GetUserByUsername(username interface{}) *user.User {
	user := user.User{}
	if db.Where("username = ?", username).Find(&user).RecordNotFound() {
		return nil
	}
	return &user
}

func GetUserByEmail(email interface{}) *user.User {
	user := user.User{}
	if db.Where("email = ?", email).Find(&user).RecordNotFound() {
		return nil
	}
	return &user
}

func GetGroup(v interface{}) *user.Group {
	group := user.Group{}
	if db.Where("id = ? OR code_name", v, v).Find(&group).RecordNotFound() {
		return nil
	}
	return &group
}

func AuthUserByPassword(username, password string) (*user.User, error) {
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

func (c Controller) Login(ctx *fiber.Ctx) {
	var err error
	var user *user.User
	var token string

	r := evo.Upgrade(ctx)
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
				"name":     user.Name,
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

func (c Controller) CreateUser(ctx *fiber.Ctx) {

	var r = evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.create.user") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var user = user.User{}
	err := r.BodyParser(&user)

	if err == nil {
		err := user.Save()
		if err == nil {
			r.WriteResponse(map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"name":     user.Name,
				"param":    user.Params,
			},
			)
		} else {
			r.WriteResponse(e.Context(err))
		}
	} else {
		r.WriteResponse(constant.ERROR_FORM_PARSE)
	}

}

func (c Controller) CreateRole(ctx *fiber.Ctx) {
	var r = evo.Upgrade(ctx)
	var role = user.Role{}
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

func (c Controller) CreateGroup(ctx *fiber.Ctx) {
	var r = evo.Upgrade(ctx)
	var group = user.Group{}
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
func (c Controller) EditUser(ctx *fiber.Ctx) {
	var r = evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.edit.user") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var user user.User
	var id = T.Must(r.Params("id")).UInt()
	if db.Where("id = ?", id).Find(&user).RecordNotFound() {
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
				"id":       user.ID,
				"username": user.Username,
				"name":     user.Name,
				"param":    user.Params,
			})
		} else {
			r.WriteResponse(err)
		}
	}

}

func (c Controller) EditRole(ctx *fiber.Ctx) {
	var r = evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.edit.role") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var role user.Role
	var id = r.Params("id")

	if db.Where("id = ? OR code_name = ?", id, id).Find(&role).RecordNotFound() {
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

func (c Controller) EditGroup(ctx *fiber.Ctx) {
	var r = evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.edit.group") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var group user.Group
	var id = r.Params("id")
	if db.Where("id = ? OR code_name = ?", id, id).Find(&group).RecordNotFound() {
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

func (c Controller) GetGroups(ctx *fiber.Ctx) {
	r := evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.group.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var groups []user.Group
	err := db.Find(&groups).Error
	if err != nil {
		r.WriteResponse(e.Field("id", constant.ERROR_INVALID_ID))
	} else {
		r.WriteResponse(groups)
	}
}

func (c Controller) GetGroup(ctx *fiber.Ctx) {
	r := evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.group.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var group user.Group
	var id = ctx.Params("id")
	if db.Where("id = ? OR code_name = ?", id, id).Find(&group).RecordNotFound() {
		r.WriteResponse(e.Field("id", constant.ERROR_INVALID_ID))
		return
	} else {
		r.WriteResponse(group)
	}
}

func (c Controller) GetRoles(ctx *fiber.Ctx) {
	r := evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.role.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var roles []user.Role
	err := db.Find(&roles).Error
	if err != nil {
		r.WriteResponse(err)
	} else {
		r.WriteResponse(roles)
	}
}

func (c Controller) GetRole(ctx *fiber.Ctx) {
	r := evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.role.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var role user.Role
	var id = ctx.Params("id")
	if db.Where("id = ? OR code_name = ?", id, id).Find(&role).RecordNotFound() {
		r.WriteResponse(e.Field("id", constant.ERROR_INVALID_ID))
		return
	} else {
		r.WriteResponse(role)
	}
}

func (c Controller) GetRoleGroups(ctx *fiber.Ctx) {
	r := evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.role.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var role user.Role
	var id = ctx.Params("id")
	if db.Where("id = ? OR code_name = ?", id, id).Find(&role).RecordNotFound() {
		r.WriteResponse(e.Field("id", constant.ERROR_INVALID_ID))
		return
	} else {
		var groups []user.Group
		db.Joins(`INNER JOIN group_roles ON "group".id = group_roles.group_id`).Where("group_roles.role_id = ?", role.ID).Find(&groups)
		r.WriteResponse(groups)
	}
}

func (c Controller) GetUser(ctx *fiber.Ctx) {
	r := evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.user.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var user user.User
	var id = ctx.Params("id")
	if db.Where("id = ? OR username = ? OR email = ?", id, id, id).Find(&user).RecordNotFound() {
		r.WriteResponse(constant.ERROR_OBJECT_NOT_EXIST)
		return
	} else {
		r.WriteResponse(user)
	}
}

func (c Controller) GetMe(ctx *fiber.Ctx) {
	r := evo.Upgrade(ctx)
	r.WriteResponse(r.User)
}

func (c Controller) GetAllUsers(ctx *fiber.Ctx) {
	r := evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.user.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var users []user.User
	err := db.Offset(ctx.Params("offset")).Limit(ctx.Params("limit")).Find(&users).Error
	if err != nil {
		r.WriteResponse(err)
	} else {
		r.WriteResponse(users)
	}
}

func (c Controller) GetAllPermissions(ctx *fiber.Ctx) {
	r := evo.Upgrade(ctx)
	if !r.User.HasPerm("auth.role.view") {
		r.WriteResponse(constant.ERROR_UNAUTHORIZED)
		return
	}
	var perms []user.Permission
	err := db.Find(&perms).Error
	if err != nil {
		r.WriteResponse(e.Context(err))
	} else {
		r.WriteResponse(perms)
	}
}
