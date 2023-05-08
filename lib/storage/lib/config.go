package lib

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type StorageSettings string

func (s StorageSettings) Parse(input string) (map[string]string, error) {
	var reg = regexp.MustCompile(`(?i)` + string(s) + `(\?(?P<querystring>.*))*`)
	var keys = reg.SubexpNames()
	var chunks = strings.Split(input, "?")
	var values = reg.FindStringSubmatch(chunks[0])
	if len(values) == 0 {
		return nil, fmt.Errorf("invalid storage settings %s, required %s", input, string(s))
	}
	d := make(map[string]string)
	for i := 1; i < len(keys); i++ {
		d[keys[i]] = values[i]
	}

	if len(chunks) == 2 {
		d["querystring"] = chunks[1]
		q, err := url.ParseQuery(chunks[1])

		if err != nil {
			return d, err
		}
		for k, _ := range q {
			d[k] = q[k][0]
		}
	}

	return d, nil
}
