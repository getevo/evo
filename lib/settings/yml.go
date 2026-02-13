package settings

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/gpath"
	"gopkg.in/yaml.v3"
	"strings"
)

// LoadYAMLSettings reads and parses a YAML file.
// It flattens nested YAML structures using dot notation.
//
// Example YAML:
//
//	database:
//	  host: localhost
//	  port: 3306
//
// Results in keys: DATABASE.HOST, DATABASE.PORT
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

	// Store both flattened and nested keys for compatibility
	flattenedMap := make(map[string]any)
	flattenMap(nestedMap, "", flattenedMap)

	for key, value := range flattenedMap {
		setData(key, value)
	}
	// Also store nested structures for Get("HTTP").Cast(&struct) pattern
	for key, value := range nestedMap {
		setData(key, value)
	}

	return nil
}

// flattenMap recursively flattens a nested map using dot notation
func flattenMap(nestedMap map[string]interface{}, parentKey string, result map[string]any) {
	for key, value := range nestedMap {
		fullKey := key
		if parentKey != "" {
			fullKey = parentKey + "." + key
		}
		switch v := value.(type) {
		case map[string]interface{}:
			flattenMap(v, fullKey, result)
		default:
			result[strings.ToUpper(fullKey)] = v
		}
	}
}

// saveYAMLSettings converts flattened settings back to nested structure and saves to YAML
func saveYAMLSettings(filename string, flattenedData map[string]any) error {
	// Convert flattened keys back to nested structure
	nested := make(map[string]interface{})

	for flatKey, value := range flattenedData {
		// Convert normalized key back to dot notation (replace _ with .)
		// and lowercase for standard YAML format
		parts := strings.Split(strings.ToLower(flatKey), "_")

		// Build nested structure
		current := nested
		for i, part := range parts[:len(parts)-1] {
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]interface{})
			}
			if next, ok := current[part].(map[string]interface{}); ok {
				current = next
			} else {
				// If we have a conflict (value exists where we need a map),
				// use the full remaining path as the key
				current[strings.Join(parts[i:], "_")] = value
				break
			}
		}

		// Set the final value
		if len(parts) > 0 {
			current[parts[len(parts)-1]] = value
		}
	}

	// Marshal to YAML
	bytes, err := yaml.Marshal(nested)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write to file
	return gpath.Write(filename, bytes)
}
