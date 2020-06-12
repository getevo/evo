package evo

import (
	"github.com/gofiber/fiber"
)

// Use registers a middleware route.
// Middleware matches requests beginning with the provided prefix.
// Providing a prefix is optional, it defaults to "/".
//
// - group.Use(handler)
// - group.Use("/api", handler)
// - group.Use("/api", handler, handler)
/*func (grp *group) Use(args ...interface{}) *fiber.Route {
	var path = ""
	var handlers []func(request *Request)
	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case string:
			path = arg
		case Handler:
			handlers = append(handlers, arg)
		default:
			log.Fatalf("Use: Invalid Handler %v", reflect.TypeOf(arg))
		}
	}
	return grp.app.register("USE", getGroupPath(grp.prefix, path), handlers...)
}*/

// Get ...
func (grp *group) Get(path string, handlers ...func(request *Request)) *fiber.Route {
	var route *fiber.Route
	for _, handler := range handlers {
		route = grp.app.Get(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
	return route
}

// Head ...
func (grp *group) Head(path string, handlers ...func(request *Request)) *fiber.Route {
	var route *fiber.Route
	for _, handler := range handlers {
		route = grp.app.Head(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
	return route
}

// Post ...
func (grp *group) Post(path string, handlers ...func(request *Request)) *fiber.Route {
	var route *fiber.Route
	for _, handler := range handlers {
		route = grp.app.Post(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
	return route
}

// Put ...
func (grp *group) Put(path string, handlers ...func(request *Request)) *fiber.Route {
	var route *fiber.Route
	for _, handler := range handlers {
		route = grp.app.Put(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
	return route
}

// Delete ...
func (grp *group) Delete(path string, handlers ...func(request *Request)) *fiber.Route {
	var route *fiber.Route
	for _, handler := range handlers {
		route = grp.app.Delete(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
	return route
}

// Connect ...
func (grp *group) Connect(path string, handlers ...func(request *Request)) *fiber.Route {
	var route *fiber.Route
	for _, handler := range handlers {
		route = grp.app.Connect(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
	return route
}

// Options ...
func (grp *group) Options(path string, handlers ...func(request *Request)) *fiber.Route {
	var route *fiber.Route
	for _, handler := range handlers {
		route = grp.app.Options(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
	return route
}

// Trace ...
func (grp *group) Trace(path string, handlers ...func(request *Request)) *fiber.Route {
	var route *fiber.Route
	for _, handler := range handlers {
		route = grp.app.Trace(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
	return route
}

// Patch ...
func (grp *group) Patch(path string, handlers ...func(request *Request)) *fiber.Route {
	var route *fiber.Route
	for _, handler := range handlers {
		route = grp.app.Patch(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
	return route
}

// All ...
func (grp *group) All(path string, handlers ...func(request *Request)) {
	for _, handler := range handlers {
		grp.app.All(path, func(ctx *fiber.Ctx) {
			r := Upgrade(ctx)
			handler(r)
		})
	}
}

// Group is used for Routes with common prefix to define a new sub-router with optional middleware.
func (grp *group) Group(prefix string, handlers ...func(request *Request)) *group {
	var route *fiber.Group
	if len(handlers) > 0 {
		for _, handler := range handlers {
			route = grp.app.Group(prefix, func(ctx *fiber.Ctx) {
				r := Upgrade(ctx)
				handler(r)
			})
		}
	} else {
		route = grp.app.Group(prefix)
	}
	gp := group{
		app: route,
	}
	return &gp
}
