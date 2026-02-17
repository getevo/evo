package settings_test

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/settings"
)

// Example demonstrates basic usage of the settings package
func Example() {
	// Initialize settings (loads from env vars, YAML, DB, CLI args)
	settings.Init()

	// Get settings with defaults
	host := settings.Get("DATABASE.HOST", "localhost").String()
	port := settings.Get("DATABASE.PORT", 3306).Int()

	fmt.Printf("Connecting to %s:%d\n", host, port)

	// Set settings
	settings.Set("APP.NAME", "MyApp")
	settings.Set("APP.VERSION", "1.0.0")

	// Output will vary based on actual config
}

// ExampleTrack demonstrates watching for setting changes
func ExampleTrack() {
	// Register callback before setting value
	settings.Track("DATABASE.HOST", func() {
		fmt.Println("Database host changed - reconnecting...")
	})

	// This will trigger the callback
	settings.Set("DATABASE.HOST", "new-host")

	// Output:
	// Database host changed - reconnecting...
}

// ExampleTrack_wildcard demonstrates wildcard pattern matching
func ExampleTrack_wildcard() {
	// Watch all database settings
	settings.Track("DATABASE.*", func() {
		fmt.Println("Database config changed - reconnecting...")
	})

	// Both will trigger the callback
	settings.Set("DATABASE.HOST", "localhost")
	settings.Set("DATABASE.PORT", 3306)

	// This won't trigger (different prefix)
	settings.Set("CACHE.HOST", "redis")

	// Output:
	// Database config changed - reconnecting...
	// Database config changed - reconnecting...
}

// ExampleTrack_all demonstrates watching all settings
func ExampleTrack_all() {
	// Watch all settings changes
	settings.Track("*", func() {
		fmt.Println("Configuration changed")
	})

	settings.Set("APP.NAME", "MyApp")
	settings.Set("DATABASE.HOST", "localhost")
	settings.Set("CACHE.ENABLED", true)

	// Output:
	// Configuration changed
	// Configuration changed
	// Configuration changed
}

// ExampleOnReload demonstrates watching for config reloads
func ExampleOnReload() {
	settings.OnReload(func() {
		fmt.Println("Configuration reloaded!")
	})

	// Reload all settings from sources
	settings.Reload()

	// Output:
	// Configuration reloaded!
}

// ExampleSet_autoPersistence demonstrates automatic database persistence
func ExampleSet_autoPersistence() {
	// When database is enabled, Set() automatically persists to database
	// Note: This example assumes db.IsEnabled() == true

	// This automatically saves to database if enabled
	settings.Set("APP.NAME", "MyApp")
	settings.Set("APP.VERSION", "1.0.0")

	// SetMulti also auto-persists
	settings.SetMulti(map[string]any{
		"DATABASE.HOST": "localhost",
		"DATABASE.PORT": 3306,
	})

	fmt.Println("Settings updated (and persisted to DB if enabled)")
	// Output:
	// Settings updated (and persisted to DB if enabled)
}

// ExampleSaveToYAML demonstrates persisting settings to YAML
func ExampleSaveToYAML() {
	// Modify some settings
	settings.Set("DATABASE.HOST", "localhost")
	settings.Set("DATABASE.PORT", 3306)
	settings.Set("APP.NAME", "MyApp")

	// Save to YAML file (always manual)
	err := settings.SaveToYAML("./config.example.yml")
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("Settings saved to YAML")
	// Output:
	// Settings saved to YAML
}

// ExampleGet_types demonstrates various type conversions
func ExampleGet_types() {
	// String
	name := settings.Get("APP.NAME", "DefaultApp").String()
	fmt.Println("Name:", name)

	// Integer
	port := settings.Get("HTTP.PORT", 8080).Int()
	fmt.Println("Port:", port)

	// Boolean
	enabled := settings.Get("FEATURE.ENABLED", true).Bool()
	fmt.Println("Enabled:", enabled)

	// Duration
	timeout, _ := settings.Get("HTTP.TIMEOUT", "30s").Duration()
	fmt.Println("Timeout:", timeout)

	// Size in bytes
	maxSize := settings.Get("UPLOAD.MAX_SIZE", "10MB").SizeInBytes()
	fmt.Println("Max size:", maxSize, "bytes")

	// Output will vary based on actual config
}

// ExampleSetMulti demonstrates setting multiple values at once
func ExampleSetMulti() {
	settings.SetMulti(map[string]any{
		"DATABASE.HOST":     "localhost",
		"DATABASE.PORT":     3306,
		"DATABASE.NAME":     "mydb",
		"DATABASE.USERNAME": "root",
	})

	fmt.Println("Multiple settings updated")
	// Output:
	// Multiple settings updated
}

// ExampleHas demonstrates checking if a setting exists
func ExampleHas() {
	settings.Set("EXISTING.KEY", "value")

	exists, value := settings.Has("EXISTING.KEY")
	fmt.Printf("Key exists: %v, Value: %s\n", exists, value.String())

	exists, _ = settings.Has("NONEXISTENT.KEY")
	fmt.Printf("Non-existent key exists: %v\n", exists)

	// Output:
	// Key exists: true, Value: value
	// Non-existent key exists: false
}
