package main

import (
	"fmt"
	"github.com/getevo/evo/cmd/evo/watcher"
	"github.com/getevo/evo/lib/cli"
	"github.com/getevo/evo/lib/gpath"
	"github.com/getevo/evo/lib/log"
	"os"
	"path/filepath"
	"time"
)

type Build struct {
	BinName     string
	BuildArgs   []string
	WorkingDir  string
	ProgramArgs []string
}

func main() {
	app := &cli.App{
		Name:  "EVO",
		Usage: "EVO manager",
		Action: func(c *cli.Context) error {
			fmt.Println("EVO Manager started")
			build()
			return nil
		},
		Commands: []*cli.Command{
			//run
			{
				Name:  "run",
				Usage: "live run application (hot reload)",
				Action: func(c *cli.Context) error {
					build()
					return nil
				},
			},
			//pack
			{
				Name:  "pack",
				Usage: "pack all needed assets to binary dir",
				Action: func(c *cli.Context) error {
					pack()
					return nil
				},
			},

			{
				Name:  "format",
				Usage: "format codes",
				Action: func(c *cli.Context) error {
					format()
					return nil
				},
			},

			{
				Name:  "admin.create",
				Usage: "evo admin.create username password",
				Action: func(c *cli.Context) error {
					adminCreate()
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func pack() {
	fmt.Println("Pack Mode")
	cfg := Build{}
	cfg.WorkingDir = gpath.WorkingDir()
	var builder = watcher.NewBuilder(cfg.WorkingDir, cfg.BinName, cfg.WorkingDir, cfg.BuildArgs)
	var runner = watcher.NewRunner(os.Stdout, os.Stderr, filepath.Join(cfg.WorkingDir, builder.Binary()), []string{"-p"})
	err := builder.Build()
	if err != nil {
		panic(err)
	}
	_, err = runner.Run()
	if err != nil {
		panic(err)
	}
	runner.Kill()

}

func build() {
	fmt.Println("Hot Reload Mode")
	cfg := Build{}
	cfg.WorkingDir = gpath.WorkingDir()
	var onBuild = false
	var builder = watcher.NewBuilder(cfg.WorkingDir, cfg.BinName, cfg.WorkingDir, cfg.BuildArgs)
	var runner = watcher.NewRunner(os.Stdout, os.Stderr, filepath.Join(cfg.WorkingDir, builder.Binary()), cfg.ProgramArgs)
	watcher.NewWatcher(cfg.WorkingDir, func() {
		if onBuild {
			fmt.Println("skip build due another build")
			return
		}
		runner.Kill()
		var counter = 0
		for runner.IsRunning() {
			counter++
			if counter > 2 {
				fmt.Println("Unable to kill process. try again ...")
				runner.Kill()
				break
			}
		}
		onBuild = true
		err := builder.Build()
		onBuild = false
		if err != nil {

			fmt.Println("\n\nBUILD FAILED:")
			log.Error(err)
		} else {

			onBuild = false
			_, err = runner.Run()
			if err != nil {
				log.Error(err)
			}
		}
	})

	onBuild = true
	err := builder.Build()
	onBuild = false
	if err != nil {
		log.Error(err)
	} else {

		_, err = runner.Run()
		if err != nil {
			fmt.Println("\n\nBUILD FAILED:")
			log.Error(err)
			//fmt.Println("")
		}
	}

	for {
		time.Sleep(1 * time.Minute)
	}

}
