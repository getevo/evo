package settings

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// LoadOSArgs loads settings from command-line arguments.
// Supports two formats:
//   - KEY=value (e.g., DATABASE.HOST=localhost)
//   - --KEY value (e.g., --DATABASE.PORT 3306)
//
// Command-line arguments have the highest priority and will override
// all other configuration sources.
func LoadOSArgs() {
	// Parse KEY=value format
	re := regexp.MustCompile(`^([a-zA-Z0-9_.]+)=(.*)$`)
	for _, arg := range os.Args {
		if matches := re.FindStringSubmatch(arg); matches != nil {
			key := matches[1]
			value := matches[2]

			// Handle quoted values
			value = unquote(value)

			setData(key, value)
		}
	}

	// Parse --KEY value format
	for idx, arg := range os.Args {
		if strings.HasPrefix(arg, "--") {
			if len(os.Args) > idx+1 {
				v := os.Args[idx+1]
				// Skip if next arg is also a flag
				if strings.HasPrefix(v, "--") {
					v = ""
				}

				v = unquote(v)
				arg = strings.TrimPrefix(arg, "--")
				setData(arg, v)
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
