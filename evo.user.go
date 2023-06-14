package evo

import "github.com/getevo/evo/v2/lib/generic"

func (r *Request) User() *UserInterface {
	if r.user == nil {
		var user = (UserInterfaceInstance).FromRequest(r)
		r.user = &user
	}
	return r.user
}

type DefaultUserInterface struct{}

type Attributes map[string]interface{}

func (d DefaultUserInterface) Attributes() Attributes {
	return Attributes{}
}

var UserInterfaceInstance UserInterface = DefaultUserInterface{}

func SetUserInterface(v UserInterface) {
	UserInterfaceInstance = v
}

type UserInterface interface {
	Name() string
	LastName() string
	FullName() string
	Email() string
	UUID() string
	ID() uint64
	Anonymous() bool
	HasPermission(permission string) bool
	Attributes() Attributes
	FromRequest(request *Request) UserInterface
}

func (d DefaultUserInterface) HasPermission(permission string) bool {
	return true
}

func (d DefaultUserInterface) Name() string {
	return ""
}

func (d DefaultUserInterface) LastName() string {
	return ""
}

func (d DefaultUserInterface) FullName() string {
	return ""
}

func (d DefaultUserInterface) Email() string {
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
