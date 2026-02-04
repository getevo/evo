package evo

import (
	"github.com/getevo/evo/v2/lib/application"
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/settings"

	dbo "github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/memo"
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

	memo.Register()

}

// Run start EVO Server
func Run() {
	Application.Run()

	//do database migrations
	if args.Exists("--migration-do") {
		dbo.TriggerOnBeforeMigration()
		err := dbo.DoMigration()
		dbo.TriggerOnAfterMigration()

		if err != nil {
			log.Error("unable to perform database migrations: ", err)
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
