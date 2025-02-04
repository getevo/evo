package settings

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/gpath"
	"gopkg.in/yaml.v3"
	"strings"
)

// LoadYAMLSettings reads and parses a YAML file into a nested map.
func LoadYAMLSettings(filename string) error {
	bytes, err := gpath.ReadFile(filename)
	if err != nil {
		return err
	}

	var nestedMap map[string]interface{}
	err = yaml.Unmarshal(bytes, &nestedMap)
	if err != nil {
		return err
	}
	flattenedMap := make(map[string]string)
	flattenMap(nestedMap, "", flattenedMap)

	for key, value := range flattenedMap {
		setData(key, value)
		//data[strings.ToUpper(key)] = value
	}
	for key, value := range nestedMap {
		setData(key, value)
		//data[strings.ToUpper(key)] = value
	}

	return nil
}

// flattenMap recursively flattens a nested map
func flattenMap(nestedMap map[string]interface{}, parentKey string, result map[string]string) {
	for key, value := range nestedMap {
		fullKey := key
		if parentKey != "" {
			fullKey = parentKey + "." + key
		}
		switch v := value.(type) {
		case map[string]interface{}:
			flattenMap(v, fullKey, result)
		default:
			result[strings.ToUpper(fullKey)] = fmt.Sprint(v)
		}
	}
}
