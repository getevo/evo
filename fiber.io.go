package evo

import (
	"github.com/gofiber/fiber"
)

// Group is used for Routes with common prefix to define a new sub-router with optional middleware.
func Group(prefix string, handlers ...func(*fiber.Ctx)) *fiber.Group {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Group(prefix, handlers...)
}

// Static append path with given prefix to static files
func Static(prefix, path string) {
	statics = append(statics, [2]string{prefix, path})
}

// Use registers a middleware route.
// Middleware matches requests beginning with the provided prefix.
// Providing a prefix is optional, it defaults to "/"
func Use(args ...interface{}) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Use(args...)
}

// Connect : https://fiber.wiki/application#http-methods
func Connect(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Connect(path, handlers...)
}

// Put : https://fiber.wiki/application#http-methods
func Put(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Put(path, handlers...)
}

// Post : https://fiber.wiki/application#http-methods
func Post(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Post(path, handlers...)
}

// Delete : https://fiber.wiki/application#http-methods
func Delete(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Delete(path, handlers...)
}

// Head : https://fiber.wiki/application#http-methods
func Head(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Head(path, handlers...)
}

// Patch : https://fiber.wiki/application#http-methods
func Patch(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Patch(path, handlers...)
}

// Options : https://fiber.wiki/application#http-methods
func Options(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Options(path, handlers...)
}

// Trace : https://fiber.wiki/application#http-methods
func Trace(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Trace(path, handlers...)
}

// Get : https://fiber.wiki/application#http-methods
func Get(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.Get(path, handlers...)
}

// All : https://fiber.wiki/application#http-methods
func All(path string, handlers ...func(*fiber.Ctx)) *fiber.App {
	if app == nil {
		panic("Access object before call Setup()")
	}
	return app.All(path, handlers...)
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
