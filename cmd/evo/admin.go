package main

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/apps/auth"
	"github.com/getevo/evo/lib/log"
	"os"
	"strings"
)

func adminCreate() {
	fmt.Println(len(os.Args))
	if len(os.Args) != 4 {
		log.Error("invalid input")
		log.Panic(strings.Join(os.Args[1:], " "))
		return
	}
	evo.Setup()
	auth.Register()
	var user = evo.User{
		GivenName: os.Args[2],
		Email:     os.Args[2],
		Username:  os.Args[2],
		Password:  os.Args[3],
		Admin:     true,
		Active:    true,
	}
	err := user.Save()
	if err != nil {
		log.Error(err)
	} else {
		log.Info("User %s with password %s created", os.Args[2], os.Args[3])
	}
}
