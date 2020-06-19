package evo

import (
	"github.com/creasty/defaults"
	"github.com/getevo/evo/lib/log"
	"github.com/getevo/go-arg"
)

type args struct {
	Config  string `arg:"-c" help:"Configuration path" default:"config.yml"`
	Pack    bool   `arg:"-p" help:"Copy assets to build dir"`
	Migrate bool   `arg:"-m" help:"Migrate Database structure"`
}

var Arg args
var argList []interface{}

// Version return app version
func (args) Version() string {
	return config.App.Name
}

func ParseArg(dest interface{}) {
	argList = append(argList, dest)
}

func parseArgs(first bool) {
	if !first {
		argList = append([]interface{}{&Arg}, argList...)
		arg.MustParse(argList...)
	} else {
		if arg.MustParse(&Arg) != nil {
			if err := defaults.Set(&Arg); err != nil {
				log.Fatal(err)
			}
		}
	}
}
