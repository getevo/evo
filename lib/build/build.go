package build

// build with -ldflags="-X 'github.com/getevo/evo/v2/lib/build.Version=v1.0.0' -X 'github.com/getevo/evo/v2/lib//build.User=$(id -u -n)' -X 'github.com/getevo/evo/v2/lib/build.Time=$(date)' -X 'github.com/getevo/evo/v2/lib/build.Commit=e21cf23'"
import "fmt"

var Version = "v0.0.1-dev"
var User = ""
var Date = ""
var Commit = ""

func Register(v, u, d, c string) {
	Version = v
	User = u
	Date = d
	Commit = c
	fmt.Println("Build", "Version:", Version, "|", "User:", User, "|", "Date:", Date, "|", "Commit:", Commit)
}
