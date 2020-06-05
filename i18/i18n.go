package i18

import "fmt"

func T(s string, params ...interface{}) string {
	return fmt.Sprintf(s, params...)
}
