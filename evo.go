package evo

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/getevo/evo/v2/lib/application"
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/getevo/evo/v2/lib/settings"

	dbo "github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/memo"
	"github.com/gofiber/fiber/v3"
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
// Returns an error if setup fails instead of calling log.Fatal, allowing for graceful error handling.
func Setup(params ...any) error {
	Application = application.GetInstance()
	var err = settings.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
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
				return fmt.Errorf("only one database driver can be registered")
			}
			dbo.RegisterDriver(d)
		}
	}

	app = fiber.New(fiberConfig)
	if settings.Get("Database.Enabled").Bool() {
		if dbo.GetDriver() == nil {
			return fmt.Errorf("Database.Enabled is true but no driver passed to evo.Setup()")
		}
		// Validate that config type matches the provided driver
		configType := strings.ToLower(settings.Get("Database.Type").String())
		driverName := dbo.GetDriver().Name()
		validNames := map[string]string{
			"mysql": "mysql", "mariadb": "mysql",
			"postgres": "postgres", "postgresql": "postgres", "pgsql": "postgres",
		}
		if expected, ok := validNames[configType]; ok && expected != driverName {
			return fmt.Errorf("Database.Type is '%s' but driver '%s' was provided", configType, driverName)
		}

		db = GetDBO()
		if db != nil {
			dbo.Register(db)
			if err := settings.LoadDatabaseSettings(); err != nil {
				log.Warning("Failed to load database settings: ", err)
				// Continue without database settings
			}
		} else {
			log.Warning("Database is nil, skipping database settings load")
		}
	}

	memo.Register()
	return nil
}

// Run start EVO Server
// Returns an error if the server fails to start, allowing for graceful shutdown.
func Run() error {
	Application.Run()

	//do database migrations
	if args.Exists("--migration-do") {

		err := dbo.DoMigration()

		if err != nil {
			log.Error("unable to perform database migrations", "error", err)
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

	// Register health check endpoints
	registerHealthCheckEndpoints()

	if Any != nil {
		app.Use(func(ctx fiber.Ctx) error {
			r := Upgrade(ctx)
			if err := Any(r); err != nil {
				return err
			}
			return nil
		})
	} else {
		// Last middleware to match anything
		app.Use(func(c fiber.Ctx) error {
			return c.SendStatus(404)
		})
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- app.Listen(http.Host + ":" + http.Port)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("unable to start web server: %w", err)
		}
		return nil
	case <-quit:
		return Shutdown()
	}
}

// GetFiber return fiber instance
func GetFiber() *fiber.App {
	return app
}

func Register(applications ...application.Application) *application.App {
	return Application.Register(applications...)
}
