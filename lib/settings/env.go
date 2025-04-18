package settings

import (
	"os"
	"strings"
)

func LoadEnvVars() {
	var variables = os.Environ()
	for _, variable := range variables {
		parts := strings.SplitN(variable, "=", 2)
		if len(parts) == 2 {
			setData(parts[0], parts[1])
		}
	}
}
