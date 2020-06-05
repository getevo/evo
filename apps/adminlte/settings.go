package adminlte

import (
	"fmt"
	"github.com/getevo/evo"
	"reflect"
	"sync"
)

type Settings struct {
	NavbarColor      string `type:"select" col:"6" label:"Navbar Color" hint:"Top navbar color" options:"blue:Blue,secondary:Gray,green:Green,cyan:Cyan,yellow:Yellow,red:Red,indigo:Indigo,navy:Navy,purple:Purple,fuchsia:Fuchsia,pink:Pink,maroon:Maroon,orange:Orange,lime:Lime,teal:Teal,olive:Olive"`
	NavbarVariation  string `type:"select" col:"6" label:"Navbar Variation" hint:"Navbar could be dark or light" options:"dark:Dark,light:Light"`
	SidebarColor     string `type:"select" col:"6" label:"Sidebar Color" hint:"Sidebar color" options:"blue:Blue,secondary:Gray,green:Green,cyan:Cyan,yellow:Yellow,red:Red,indigo:Indigo,navy:Navy,purple:Purple,fuchsia:Fuchsia,pink:Pink,maroon:Maroon,orange:Orange,lime:Lime,teal:Teal,olive:Olive"`
	SidebarVariation string `type:"select" col:"6" label:"Sidebar Variation" hint:"Sidebar could be dark or light" options:"dark:Dark,light:Light"`
	AccentColor      string `type:"select" col:"6" label:"Accent Color" hint:"Heading text color" options:"blue:Blue,secondary:Gray,green:Green,cyan:Cyan,yellow:Yellow,red:Red,indigo:Indigo,navy:Navy,purple:Purple,fuchsia:Fuchsia,pink:Pink,maroon:Maroon,orange:Orange,lime:Lime,teal:Teal,olive:Olive"`
	mu               sync.Mutex
}

func (settings Settings) Get(key string) interface{} {
	settings.mu.Lock()
	defer settings.mu.Unlock()
	ref := reflect.ValueOf(settings)
	return ref.FieldByName(key).Interface()
}

func (s *Settings) OnUpdate(r *evo.Request) bool {
	fmt.Println("on update called")
	return true
}
