package settings

import (
	"github.com/getevo/evo/v2/lib/dot"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func LoadOSArgs() {
	//parse X.Y=value
	re := regexp.MustCompile(`^([a-zA-Z0-9_.]+)=(.*)$`)
	for _, arg := range os.Args {
		if matches := re.FindStringSubmatch(arg); matches != nil {
			key := matches[1]
			value := matches[2]

			// Handle quoted values
			value = unquote(value)

			data[strings.ToUpper(key)] = value
			_ = dot.Set(&data, key, value)
		}
	}

	// parse -X.Y value
	for idx, arg := range os.Args {
		if strings.HasPrefix(arg, "--") {

			if len(os.Args) > idx+1 {
				v := os.Args[idx+1]
				if strings.HasPrefix(v, "--") {
					v = ""
				}

				v = unquote(v)
				arg = strings.TrimPrefix(arg, "--")
				data[strings.ToUpper(arg)] = v
				_ = dot.Set(&data, arg, v)
			}
		}
	}
}

func unquote(v string) string {
	if (strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`)) ||
		(strings.HasPrefix(v, `'`) && strings.HasSuffix(v, `'`)) ||
		(strings.HasPrefix(v, "`") && strings.HasSuffix(v, "`")) {
		unquotedValue, _ := strconv.Unquote(v)
		v = unquotedValue
	}
	return v
}
