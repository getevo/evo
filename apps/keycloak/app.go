package keycloak

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/lib/log"
	"github.com/getevo/evo/menu"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/json"
	"io/ioutil"
	"net/http"
)

var Server string
var Realm string
var Client string
var Certificates jose.JSONWebKeySet

func Register(server, realm, client string) {
	Server = server
	Realm = realm
	Client = client
	evo.Register(App{})

}

type App struct{}

// Register the bible
func (App) Register() {
	fmt.Println("Keycloak Registered")
	resp, err := http.Get(fmt.Sprintf("%s/auth/realms/%s/protocol/openid-connect/certs", Server, Realm))
	if err != nil {
		log.Error("Unable connect to keycloak server")
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Error("Unable connect parse keycloak cert")
		log.Fatal(err)
	}
	Certificates = jose.JSONWebKeySet{}
	err = json.Unmarshal(body, &Certificates)
	if err != nil {
		log.Error("Unable connect parse keycloak cert")
		log.Fatal(err)
	}
	evo.SetUserInterface(User{})

}

// WhenReady called after setup all apps
func (App) WhenReady() {}

// Router setup routers
func (App) Router() {
	evo.Get("me", func(request *evo.Request) {
		request.WriteResponse(request.User)
	})
}

// Permissions setup permissions of app
func (App) Permissions() []evo.Permission { return []evo.Permission{} }

// Menus setup menus
func (App) Menus() []menu.Menu {
	return []menu.Menu{}
}

func (App) Pack() {}
