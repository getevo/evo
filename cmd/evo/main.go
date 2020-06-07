package main

import (
	"fmt"
	"github.com/getevo/evo/cmd/evo/watcher"
	"github.com/getevo/evo/lib/gpath"
	"github.com/getevo/evo/lib/log"
	"github.com/urfave/cli"
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
		fmt.Println("Change detected")
		if onBuild {
			fmt.Println("skip build due another build")
			return
		}
		runner.Kill()
		var counter = 0
		for runner.IsRunning() {
			time.Sleep(1 * time.Second)
			counter++
			if counter > 2 {
				fmt.Println("Unable to kill process. try again ...")
				runner.Kill()
				return
			}
		}
		onBuild = true
		err := builder.Build()
		if err != nil {
			panic(err)
		}

		onBuild = false
		_, err = runner.Run()
		if err != nil {
			panic(err)
		}
	})

	onBuild = true
	err := builder.Build()
	if err != nil {
		panic(err)
	}

	onBuild = false
	_, err = runner.Run()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(1 * time.Minute)
	}

}
