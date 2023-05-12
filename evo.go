package evo

import (
	"github.com/getevo/evo/v2/lib/cache"
	dbo "github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/settings"
	"github.com/getevo/evo/v2/lib/settings/database"
	"github.com/gofiber/fiber/v2"
	"log"
)

var (
	app *fiber.App
	Any func(request *Request) error
)
var http = HTTPConfig{}
var fiberConfig = fiber.Config{}

// Setup set up the EVO app
func Setup() {
	var err = settings.Init()
	if err != nil {
		log.Fatal(err)
	}
	settings.Register("HTTP", &http)

	err = generic.Parse(http).Cast(&fiberConfig)

	app = fiber.New(fiberConfig)
	if settings.Get("Database.Enabled").Bool() {
		database.SetDBO(GetDBO())
		settings.SetDefaultDriver(database.Driver)
		dbo.Register()
	}

	cache.Register()
}

// Run start EVO Server
func Run() {
	if Any != nil {
		app.Use(func(ctx *fiber.Ctx) error {
			r := Upgrade(ctx)
			if err := Any(r); err != nil {
				return err
			}
			return nil
		})
	} else {
		// Last middleware to match anything
		app.Use(func(c *fiber.Ctx) error {
			c.SendStatus(404)
			return nil
		})
	}

	var err error
	err = app.Listen(http.Host + ":" + http.Port)

	log.Fatal("unable to start web server", "error", err)
}

// GetFiber return fiber instance
func GetFiber() *fiber.App {
	return app
}
