package websocket

import (
	"github.com/getevo/evo"
	"github.com/getevo/evo/menu"
	"github.com/gofiber/fiber"
	"github.com/gofiber/websocket"
)

func Register() {
	evo.Register(App{})
}

type App struct{}

func (App) Register() {

}

func (App) Router() {
	app := evo.GetFiber()
	app.Use(func(c *fiber.Ctx) {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			c.Next()
		}
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// c.Locals is added to the *websocket.Conn
		/*	fmt.Println(c.Locals("allowed"))  // true

			// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
			for {
				mt, msg, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					break
				}
				log.Printf("recv: %s", msg)
				err = c.WriteMessage(mt, msg)
				if err != nil {
					log.Println("write:", err)
					break
				}
			}*/

	}))
}

func (App) Permissions() []evo.Permission {
	return []evo.Permission{}
}

func (App) Menus() []menu.Menu {
	return []menu.Menu{}
}
func (App) WhenReady() {}

func (App) Pack() {}
