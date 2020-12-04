package evo

import (
	"crypto/tls"
	"fmt"
	"github.com/AlexanderGrom/go-event"
	"github.com/getevo/evo/lib/gpath"
	"github.com/getevo/evo/lib/jwt"
	"github.com/getevo/evo/lib/log"
	"github.com/getevo/evo/lib/text"
	"github.com/getevo/evo/lib/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	recovermd "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	//Public
	app *fiber.App

	Events          = event.New()
	StatusCodePages = map[int]string{}

	//private
	statics [][2]string
)

// Setup setup the EVO app
func Setup() {
	parseArgs(true)
	fmt.Printf("Input args %+v \n", Arg)

	parseConfig()
	if Arg.Pack {
		config.Database.Debug = "false"
	}
	bodySize, err := utils.ParseSize(config.Server.MaxUploadSize)
	if err != nil {
		bodySize = 10 * 1024 * 1024
	}

	app = fiber.New(fiber.Config{
		Prefork:       config.Tweaks.PreFork,
		StrictRouting: config.Server.StrictRouting,
		CaseSensitive: config.Server.CaseSensitive,
		ServerHeader:  config.Server.Name,
		BodyLimit:     int(bodySize),
	})

	if config.CORS.Enabled {
		fmt.Println("Enabled CORS Middleware")
		CORS := config.CORS
		c := cors.Config{
			AllowCredentials: CORS.AllowCredentials,
			AllowHeaders:     strings.Join(CORS.AllowHeaders, ","),
			AllowMethods:     strings.Join(CORS.AllowMethods, ","),
			AllowOrigins:     CORS.AllowOrigins,
			MaxAge:           CORS.MaxAge,
		}
		app.Use(cors.New(c))
	}

	if config.RateLimit.Enabled {
		fmt.Println("Enabled Rate Limiter")
		cfg := limiter.Config{
			Expiration: time.Duration(config.RateLimit.Duration) * time.Second,
			Max:        config.RateLimit.Requests,
		}
		app.Use(limiter.New(cfg))
	}

	if config.Server.Debug {
		fmt.Println("Enabled Logger")
		app.Use(logger.New())
		if config.Server.Recover {
			app.Use(recovermd.New())
		}

		//app.Use("/swagger", swagger.Handler) // default
	} else {
		if config.Server.Recover {
			app.Use(recovermd.New())
		}
	}

	if config.Server.RequestID {
		fmt.Println("Enabled Request ID")
		app.Use(requestid.New())
	}

	Static("/", config.App.Static)
	//app.Settings.TemplateEngine = template.Handlebars()

	jwt.Register(text.ToJSON(config.JWT))
	if config.Database.Enabled {
		GetDBO()
		InitUserModel(Database, config)
	}

}

// CustomError set custom page for errors
func CustomError(code int, path string) error {
	if gpath.IsFileExist(path) {
		StatusCodePages[code] = path
		return nil
	} else if gpath.IsFileExist(config.App.Static + "/" + path) {
		StatusCodePages[code] = config.App.Static + "/" + path
		return nil
	}
	return fmt.Errorf("custom error page %d not found %s", code, path)

}

// Run start EVO Server
func Run() {

	if Arg.Pack {
		return
	}
	parseArgs(false)

	//Static Files
	for _, item := range statics {
		app.Static(item[0], item[1])
	}

	// Last middleware to match anything
	app.Use(func(c *fiber.Ctx) error {
		if file, ok := StatusCodePages[404]; c.Method() == "GET" && ok {
			c.SendFile(config.App.Static + "/" + file)
		}
		c.SendStatus(404)
		return nil
	})
	Events.Go("init.after")

	for _, item := range onReady {
		item()
	}

	var err error
	if config.Server.HTTPS {
		cer, err := tls.LoadX509KeyPair(GuessPath(config.Server.Cert), GuessPath(config.Server.Key))
		if err != nil {
			log.Fatal(err)
		}
		//err = app.Listen(config.Server.Host+":"+config.Server.Port, &tls.Config{Certificates: []tls.Certificate{cer}})
		ln, _ := net.Listen("tcp", config.Server.Host+":"+config.Server.Port)
		ln = tls.NewListener(ln, &tls.Config{Certificates: []tls.Certificate{cer}})
		err = app.Listen(config.Server.Host + ":" + config.Server.Port)
	} else {
		err = app.Listen(config.Server.Host + ":" + config.Server.Port)
	}
	Events.Go("server.panic")
	log.Fatal(err)
}

// GetFiber return fiber instance
func GetFiber() *fiber.App {
	return app
}

func Pack(path string) {
	name := filepath.Base(path)
	WorkingDir = gpath.WorkingDir()
	dest := WorkingDir + "/bundle/" + name

	err := gpath.MakePath(dest)
	if err != nil {
		log.Fatal(err)
	}
	f, err := gpath.Open(WorkingDir + "/bundle/.ignore")
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString("#evo")
	len := len(path)
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			gpath.MakePath(dest + p[len:])
		} else {
			if !strings.HasSuffix(info.Name(), ".go") {
				gpath.CopyFile(p, dest+p[len:])
			}
		}
		return nil
	})

}
