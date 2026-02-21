package evo

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// Get ...
func (grp *group) Get(path string, handlers ...Handler) fiber.Router {
	var route fiber.Router

	route = (*grp.app).Get(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Head ...
func (grp *group) Head(path string, handlers ...Handler) fiber.Router {
	var route fiber.Router
	route = (*grp.app).Head(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Post ...
func (grp *group) Post(path string, handlers ...Handler) fiber.Router {
	var route fiber.Router
	route = (*grp.app).Post(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})
	return route
}

// Put ...
func (grp *group) Put(path string, handlers ...Handler) fiber.Router {
	var route fiber.Router
	route = (*grp.app).Put(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})
	return route
}

// Delete ...
func (grp *group) Delete(path string, handlers ...Handler) fiber.Router {
	var route fiber.Router
	route = (*grp.app).Delete(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})
	return route
}

// Connect ...
func (grp *group) Connect(path string, handlers ...Handler) fiber.Router {
	var route fiber.Router
	route = (*grp.app).Connect(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})
	return route
}

// Options ...
func (grp *group) Options(path string, handlers ...Handler) fiber.Router {
	var route fiber.Router
	route = (*grp.app).Options(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Trace ...
func (grp *group) Trace(path string, handlers ...Handler) fiber.Router {
	var route fiber.Router

	route = (*grp.app).Trace(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})
	return route
}

// Patch ...
func (grp *group) Patch(path string, handlers ...Handler) fiber.Router {
	var route fiber.Router
	route = (*grp.app).Patch(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// All ...
func (grp *group) All(path string, handlers ...Handler) {

	(*grp.app).All(path, func(ctx fiber.Ctx) error {
		return handle(ctx, handlers)
	})

}

// Group is used for Routes with common prefix to define a new sub-router with optional middleware.
func (grp *group) Group(prefix string, handlers ...Handler) group {
	var route fiber.Router
	if len(handlers) > 0 {
		route = (*grp.app).Group(prefix, func(ctx fiber.Ctx) error {
			return handle(ctx, handlers)
		})
	} else {
		route = (*grp.app).Group(prefix)
	}
	gp := group{
		app: &route,
	}
	return gp
}

// Static serves files from the file system
func (grp *group) Static(path string, dir string, config ...static.Config) group {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var cfg static.Config
	if len(config) > 0 {
		cfg = config[0]
	}
	var route fiber.Router
	route = (*grp.app).Use(path, static.New(dir, cfg))
	return group{app: &route}
}
