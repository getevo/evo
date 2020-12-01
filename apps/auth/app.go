// @doc type 		app
// @doc name		auth
// @doc description authentication api
// @doc author		reza
// @doc include		github.com/getevo/evo/user.model.go
package auth

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/apps/query"
	"github.com/getevo/evo/menu"
	"gorm.io/gorm"
)

// Register register app
func Register() {
	evo.Register(App{})
}

var db *gorm.DB
var config *evo.Configuration

// App settings app struct
type App struct{}

// Register register the auth in io apps
func (App) Register() {

	fmt.Println("Auth Registered")

	userFilter := query.Filter{
		Object: &evo.User{},
		Slug:   "user",
		Allowed: map[string]string{
			"id":         ``,
			"username":   `validate:"format=username"`,
			"name":       `validate:"format=text"`,
			"email":      `validate:"format=text"`,
			"group_id":   `validate:"format=numeric"`,
			"created_at": `validate:"format=date"`,
		},
	}
	//userFilter.SetFilter("admin = ?",false)
	query.Register(userFilter)

	db = evo.GetDBO()
	config = evo.GetConfig()
	if config.Database.Enabled == false {
		panic("Auth App require database to be enabled. solution: enable database at config.yml")
	}
}

// Router setup routers
func (App) Router() {
	controller := Controller{}

	// @doc type 		meta
	// @doc prefix		/auth
	var auth = evo.Group("/auth")

	// @doc type 		api
	// @doc name 		login
	// @doc description Possible return json or set cookie
	// @doc body   		Model:#auth.AuthParams
	// @doc return		Note:json cookie or text based on accept parameter
	// @doc required	Header:accept(text/html,application/json)
	auth.Post("/user/login", controller.Login)

	// @doc type 		api
	// @doc name 		user create
	// @doc body   		Model:#evo.User
	// @doc return 		Map:id,username,name,param
	// @doc required	Permission:#auth.user.create
	auth.Post("/user/create", controller.CreateUser)

	// @doc type 		api
	// @doc name 		user edit
	// @doc describe    :id (user id)
	// @doc body   		Model:#evo.User
	// @doc return 		Map:id,username,name,param
	// @doc required	Permission:#auth.user.edit
	auth.Post("/user/edit/:id", controller.EditUser)

	// @doc type 		api
	// @doc name 		me
	// @doc description return current logged in user
	// @doc path   		/user/me
	// @doc return 		Model:#evo.User
	auth.Get("/user/me", controller.GetMe)

	// @doc type 		api
	// @doc name 		list users
	// @doc description return list of users for given offset and range
	// @doc describe    :limit (limit number of showing users)
	// @doc describe    :offset (starting offset)
	// @doc return 		Model:[]evo.User
	// @doc required	Permission:#auth.user.view
	auth.Get("/user/all/:offset/:limit", controller.GetAllUsers)

	// @doc type 		api
	// @doc name 		get user
	// @doc description get single user by id
	// @doc describe    :id (user id)
	// @doc return 		Model:#evo.User
	// @doc required	Permission:#auth.user.view
	auth.Get("/user/:id", controller.GetUser) //this should be always last router else it will match before others

	// @doc type 		api
	// @doc name 		create role
	// @doc body   		Model:#evo.Role
	// @doc return 		Map:id,name,code_name,parent
	// @doc required	Permission:#auth.role.create
	auth.Post("/role/create", controller.CreateRole)

	// @doc type 		api
	// @doc name 		edit role
	// @doc describe    :id (role id)
	// @doc body   		Model:#evo.Role
	// @doc return 		Map:id,name,code_name,parent
	// @doc required	Permission:#auth.user.edit
	auth.Post("/role/edit/:id", controller.EditRole)

	// @doc type 		api
	// @doc name 		roles
	// @doc description get list of all roles
	// @doc return 		Model:#[]user.Role
	// @doc required	Permission:#auth.role.view
	auth.Get("/role/all", controller.GetRoles)

	// @doc type 		api
	// @doc name 		get role
	// @doc describe    :id (role id)
	// @doc description get single role
	// @doc return 		Model:#evo.Role
	// @doc required	Permission:#auth.role.view
	auth.Get("/role/:id", controller.GetRole)

	// @doc type 		api
	// @doc name 		roles
	// @doc description get list of all groups of a single role
	// @doc describe    :id (role id)
	// @doc return 		Model:#[]user.Role
	// @doc required	Permission:#auth.role.view
	auth.Get("/role/:id/groups", controller.GetRoleGroups)

	// @doc type 		api
	// @doc name 		group create
	// @doc return 		Model:#evo.Group
	// @doc body   		Model:#evo.Group
	// @doc required	Permission:#auth.group.create
	auth.Post("/group/create", controller.CreateGroup)

	// @doc type 		api
	// @doc name 		group edit
	// @doc describe    :id (group id)
	// @doc return 		Model:#evo.Group
	// @doc body   		Model:#evo.Group
	// @doc required	Permission:#auth.group.create
	auth.Post("/group/edit/:id", controller.EditGroup)

	// @doc type 		api
	// @doc name 		get groups
	// @doc return 		Model:#[]user.Group
	// @doc required	Permission:#auth.group.view
	auth.Get("/group/all", controller.GetGroups)

	// @doc type 		api
	// @doc name 		get single group
	// @doc describe    :id (group id)
	// @doc return 		Model:#[]user.Group
	// @doc required	Permission:#auth.group.view
	auth.Get("/group/:id", controller.GetGroup)

	// @doc type 		api
	// @doc name 		all permissions
	// @doc description get list of all system permissions
	// @doc return 		Model:#[]user.Permission
	// @doc required	Permission:#auth.role.view
	auth.Get("/permission/all", controller.GetAllPermissions)

	// @doc group
}

// Permissions setup permissions of app
func (App) Permissions() []evo.Permission {
	// @doc type permission
	return []evo.Permission{
		{Title: "Access Users", CodeName: "user.view", Description: "Access list to view list of users"},
		{Title: "Create Users", CodeName: "user.create", Description: "Create new user"},
		{Title: "Edit users", CodeName: "user.edit", Description: "Edit user data"},
		{Title: "Remove users", CodeName: "user.remove", Description: "Remove user data"},
		{Title: "Login as user", CodeName: "user.loginas", Description: "Login as another user without credentials"},
		{Title: "Limit to subgroup", CodeName: "user.limited", Description: "User can only access/modify subgroup users"},
		{Title: "Access Groups", CodeName: "group.view", Description: "Access to groups data"},
		{Title: "Create Groups", CodeName: "group.create", Description: "Create new group"},
		{Title: "Edit Groups", CodeName: "group.edit", Description: "Edit groups data"},
		{Title: "Remove Groups", CodeName: "group.remove", Description: "Remove groups"},
		{Title: "Access Roles", CodeName: "role.view", Description: "Access to roles"},
		{Title: "Create Role", CodeName: "role.create", Description: "Create new role"},
		{Title: "Edit Roles", CodeName: "role.edit", Description: "Edit roles data"},
		{Title: "Remove Roles", CodeName: "role.remove", Description: "Remove roles"},
	}
}

// Menus setup menus
func (App) Menus() []menu.Menu {
	return []menu.Menu{}
}

// WhenReady called after setup all apps
func (App) WhenReady() {}

func (App) Pack() {}
