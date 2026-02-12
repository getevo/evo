package evo

import (
	"os"
	"strings"

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

// Setup set up the EVO app.
// Optional params: pass a db.Driver to select the database driver (e.g. pgsql.Driver{} or mysql.Driver{}).
func Setup(params ...any) {
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

	// Extract driver from params
	var driverCount int
	for _, p := range params {
		if d, ok := p.(dbo.Driver); ok {
			driverCount++
			if driverCount > 1 {
				log.Fatal("only one database driver can be registered")
			}
			dbo.RegisterDriver(d)
		}
	}

	app = fiber.New(fiberConfig)
	if settings.Get("Database.Enabled").Bool() {
		if dbo.GetDriver() == nil {
			log.Fatal("Database.Enabled is true but no driver passed to evo.Setup()")
		}
		// Validate that config type matches the provided driver
		configType := strings.ToLower(settings.Get("Database.Type").String())
		driverName := dbo.GetDriver().Name()
		validNames := map[string]string{
			"mysql": "mysql", "mariadb": "mysql",
			"postgres": "postgres", "postgresql": "postgres", "pgsql": "postgres",
		}
		if expected, ok := validNames[configType]; ok && expected != driverName {
			log.Fatal("Database.Type is '", configType, "' but driver '", driverName, "' was provided")
		}

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

		err := dbo.DoMigration()

		if err != nil {
			log.Error("unable to perform database migrations: ", err)
		} else {
			log.Info("database migrations performed successfully")
		}
	}

	if args.Exists("--migration-dry-run") {
		dbo.DryRunMigration()
		os.Exit(0)
	}

	if args.Exists("--migration-dump") {
		dbo.DumpSchema()
		os.Exit(0)
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
