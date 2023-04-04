package settings

import (
	"github.com/gofiber/fiber/v2"
	"testing"
)

type MyStruct struct {
	ServerHeader  string `json:"server_header"`
	StrictRouting bool   `json:"strict_routing"`
}

func TestConfigParser(t *testing.T) {
	Init("test.yml")
	/*var cfg = fiber.Config{}
	g := generic.Parse(&cfg)
	Init("test.yml")
	for _, field := range g.Props() {
		if field.Tag.Get("json") != "-" {
			if field.Tag.Get("json") != "" {
				Register(Setting{
					Domain: "HTTP",
					Name:   field.Tag.Get("json"),
					Value:  g.Prop(field.Name).String(),
				})
				g.SetProp(field.Name, Get("HTTP."+field.Tag.Get("json")))
			}
		}
	}
	*/

	var st = fiber.Config{
		ServerHeader: "My Default Server Header",
	}

	Register("HTTP", &st)

}
