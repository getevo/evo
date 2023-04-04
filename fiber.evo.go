package evo

import (
	"github.com/gofiber/fiber/v2"
)

type Handler func(request *Request) interface{}
type Middleware func(request *Request) error

type group struct {
	app *fiber.Router
}

func (grp *group) Name(name string) *group {
	var router = (*grp.app).Name(name + ".")
	grp.app = &router
	return grp
}

// Group is used for Routes with common prefix to define a new sub-router with optional middleware.
func Group(path string, handlers ...Middleware) *group {
	if app == nil {
		panic("Access object before call Setup()")
	}

	var route fiber.Router
	if len(handlers) > 0 {
		route = app.Group(path, func(ctx *fiber.Ctx) error {
			r := Upgrade(ctx)
			for _, handler := range handlers {
				var response = handler(r)
				if response != nil {
					r.WriteResponse(response)
					break
				}
			}

			return nil
		})
	} else {
		route = app.Group(path)
	}

	gp := group{
		app: &route,
	}
	return &gp
}

/*func Group(prefix string, handlers ...func(*fiber.Ctx)) *fiber.Group {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Group(prefix, handlers...)
}*/

// Use registers a middleware route.
// Middleware matches requests beginning with the provided prefix.
// Providing a prefix is optional, it defaults to "/"
func Use(path string, middleware Middleware) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Use(path, func(ctx *fiber.Ctx) error {
		r := Upgrade(ctx)
		if err := middleware(r); err != nil {
			return err
		}

		return nil
	})

	return route
}

// Connect : https://fiber.wiki/application#http-methods
func Connect(path string, handlers ...Handler) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Connect(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Put : https://fiber.wiki/application#http-methods
func Put(path string, handlers ...Handler) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Put(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Post : https://fiber.wiki/application#http-methods
func Post(path string, handlers ...Handler) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Post(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

func handle(ctx *fiber.Ctx, handlers []Handler) error {
	r := Upgrade(ctx)
	for _, handler := range handlers {
		var resp = handler(r)
		if r._break {
			return nil
		}
		if resp != nil {
			r.WriteResponse(resp)
			break
		}
	}
	return nil
}

// Delete : https://fiber.wiki/application#http-methods
func Delete(path string, handlers ...Handler) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Delete(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Head : https://fiber.wiki/application#http-methods
func Head(path string, handlers ...Handler) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Head(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Patch : https://fiber.wiki/application#http-methods
func Patch(path string, handlers ...Handler) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Patch(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Options : https://fiber.wiki/application#http-methods
func Options(path string, handlers ...Handler) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Options(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Trace : https://fiber.wiki/application#http-methods
func Trace(path string, handlers ...Handler) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Trace(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// Get : https://fiber.wiki/application#http-methods
func Get(path string, handlers ...Handler) fiber.Router {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var route fiber.Router
	route = app.Get(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

	return route
}

// All : https://fiber.wiki/application#http-methods
func All(path string, handlers ...Handler) {
	if app == nil {
		panic("Access object before call Setup()")
	}

	app.All(path, func(ctx *fiber.Ctx) error {
		return handle(ctx, handlers)
	})

}

// Shutdown gracefully shuts down the server without interrupting any active connections.
// Shutdown works by first closing all open listeners and then waiting indefinitely for all connections to return to idle and then shut down.
//
// When Shutdown is called, Serve, ListenAndServe, and ListenAndServeTLS immediately return nil.
// Make sure the program doesn't exit and waits instead for Shutdown to return.
//
// Shutdown does not close keepalive connections so its recommended to set ReadTimeout to something else than 0.
func Shutdown() error {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Shutdown()
}

// Redirect redirects a path to another
func Redirect(path, to string, code ...int) {
	if app == nil {
		panic("Access object before call Setup()")
	}
	var redirectCode = fiber.StatusTemporaryRedirect
	if len(code) > 0 {
		redirectCode = code[0]
	}
	app.All(path, func(ctx *fiber.Ctx) error {
		return ctx.Redirect(to, redirectCode)
	})
}

// RedirectPermanent redirects a path to another with code 301
func RedirectPermanent(path, to string) {
	Redirect(path, to, 301)
}

// RedirectTemporary redirects a path to another with code 302
func RedirectTemporary(path, to string) {
	Redirect(path, to, 302)
}
