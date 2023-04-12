package build

// build with -ldflags="-X 'github.com/getevo/evo/v2/lib/build.Version=v1.0.0' -X 'github.com/getevo/evo/v2/lib//build.User=$(id -u -n)' -X 'github.com/getevo/evo/v2/lib/build.Time=$(date)' -X 'github.com/getevo/evo/v2/lib/build.Commit=e21cf23'"
import (
	"fmt"
)

var Version = "v0.0.1-dev"
var User = ""
var Date = ""
var Commit = ""

func Register() {
	fmt.Println("Build", "Version:", Version, "|", "User:", User, "|", "Date:", Date, "|", "Commit:", Commit)
	info.Version = Version
	info.User = User
	info.Date = Date
	info.Commit = Commit
}

type Information struct {
	Version string `json:"version"`
	User    string `json:"user"`
	Date    string `json:"date"`
	Commit  string `json:"commit"`
}

var info = Information{}

func GetInfo() Information {
	return info
}
