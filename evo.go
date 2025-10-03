package evo

import (
	"github.com/getevo/evo/v2/lib/application"
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/settings"
	"github.com/gofiber/fiber/v2/middleware/logger"

	dbo "github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/gofiber/fiber/v2"
)

var (
	app *fiber.App
	Any func(request *Request) error
)
var http = HTTPConfig{}
var fiberConfig = fiber.Config{}
var Application *application.App

// Setup set up the EVO app
func Setup() {
	Application = application.GetInstance()
	var err = settings.Init()
	if err != nil {
		log.Fatal(err)
	}
	if settings.Get("LOG.LEVEL").String() != "" {
		log.SetLevel(log.ParseLevel(settings.Get("LOG.LEVEL").String()))
	}
	if args.Exists("--debug") {
		log.SetLevel(log.DebugLevel)
	}
	settings.Register("HTTP", &http)
	settings.Get("HTTP").Cast(&http)
	err = generic.Parse(http).Cast(&fiberConfig)
	if err != nil {
		log.Fatal("Unable to retrieve HTTP server configurations: ", err)
	}

	app = fiber.New(fiberConfig)
	if settings.Get("Database.Enabled").Bool() {
		db = GetDBO()
		dbo.Register(db)
		settings.LoadDatabaseSettings()
	}
	if settings.Get("HTTP.PrintRequest").Bool() {
		GetFiber().Use(logger.New(logger.Config{
			Next: func(c *fiber.Ctx) bool {
				var path = c.Path()
				if path == "/health" {
					return true
				}
				return false
			},
		}))
	}
	Get("/health", func(request *Request) any {
		return "ok"
	})
}

// Run start EVO Server
func Run() {
	Application.Run()

	//do database migrations
	if args.Exists("--migration-do") {
		err := dbo.DoMigration()
		if err != nil {
			log.Fatal("unable to perform database migrations: ", err)
		} else {
			log.Info("database migrations performed successfully")
		}
	}

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

func Register(applications ...application.Application) *application.App {
	return Application.Register(applications...)
}
