package evo

import "github.com/getevo/evo/v2/lib/generic"

func (r *Request) User() UserInterface {
	if r.UserInterface == nil {
		var user = (UserInterfaceInstance).FromRequest(r)
		if user == nil {
			user = DefaultUserInterface{}
		}
		r.UserInterface = &user
	}
	return *r.UserInterface
}

type DefaultUserInterface struct{}

type Attributes map[string]any

func (d DefaultUserInterface) Attributes() Attributes {
	return Attributes{}
}

var UserInterfaceInstance UserInterface = DefaultUserInterface{}

func SetUserInterface(v UserInterface) {
	UserInterfaceInstance = v
}

type UserInterface interface {
	GetFirstName() string
	GetLastName() string
	GetFullName() string
	GetEmail() string
	UUID() string
	ID() uint64
	Anonymous() bool
	HasPermission(permission string) bool
	Attributes() Attributes
	Interface() interface{}
	FromRequest(request *Request) UserInterface
}

func (d DefaultUserInterface) HasPermission(permission string) bool {
	return false
}

func (d DefaultUserInterface) GetFirstName() string {
	return ""
}

func (d DefaultUserInterface) GetLastName() string {
	return ""
}

func (d DefaultUserInterface) GetFullName() string {
	return ""
}

func (d DefaultUserInterface) GetEmail() string {
	return ""
}

func (d DefaultUserInterface) UUID() string {
	return ""
}

func (d DefaultUserInterface) ID() uint64 {
	return 0
}

func (d DefaultUserInterface) Anonymous() bool {
	return true
}

func (d DefaultUserInterface) Interface() interface{} {
	return d
}

func (d DefaultUserInterface) FromRequest(request *Request) UserInterface {
	return DefaultUserInterface{}
}

func (a Attributes) Has(key string) bool {
	_, ok := a[key]
	return ok
}

func (a Attributes) Get(key string) generic.Value {
	v, _ := a[key]
	return generic.Parse(v)
}
