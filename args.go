package evo

import "github.com/alexflint/go-arg"

type args struct {
	Config string `arg:"env" help:"Configuration path" default:"config.yml"`
}

var Arg args

// Version return app version
func (args) Version() string {
	return config.App.Name
}
func parseArgs() {
	arg.MustParse(&Arg)

}
