package evo

import (
	"github.com/alexflint/go-arg"
)

type args struct {
	Config string `arg:"env" help:"Configuration path" default:"config.yml"`
	Pack   bool   `arg:"-p" help:"Copy assets to build dir"`
}

var Arg args

// Version return app version
func (args) Version() string {
	return config.App.Name
}
func parseArgs() {
	// TODO: Dynamic args
	arg.Parse(&Arg)
	if Arg.Config == "" {
		Arg.Config = "config.yml"
	}

}
