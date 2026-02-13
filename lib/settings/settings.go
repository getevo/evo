// Package settings provides a thread-safe, multi-source configuration management system.
// It supports loading settings from environment variables, YAML files, databases, and command-line arguments
// with automatic type conversion and hierarchical key support using dot notation.
//
// # Loading Priority (highest to lowest):
//  1. Command-line arguments (--key=value)
//  2. YAML configuration file
//  3. Database settings (if enabled)
//  4. Environment variables
//
// # Example Usage:
//
//	// Get a setting with default value
//	host := settings.Get("DATABASE.HOST", "localhost").String()
//	port := settings.Get("DATABASE.PORT", 3306).Int()
//	enabled := settings.Get("FEATURE.ENABLED", false).Bool()
//
//	// Set a setting
//	settings.Set("APP.NAME", "MyApp")
//
//	// Watch for changes
//	settings.OnReload(func() {
//	    log.Info("Settings reloaded")
//	})
//
//	// Watch specific setting
//	settings.OnChange("DATABASE.HOST", func() {
//	    log.Info("Database host changed - reconnecting...")
//	    db.Reconnect()
//	})
//
//	// Watch all database settings (wildcard)
//	settings.OnChange("DATABASE.*", func() {
//	    log.Info("Database setting changed - reconnecting...")
//	    db.Reconnect()
//	})
//
//	// Watch all settings
//	settings.OnChange("*", func() {
//	    log.Info("Configuration changed - reloading...")
//	    app.ReloadConfig()
//	})
//
//	// Persist changes
//	settings.SaveToYAML("./config.yml")
package settings

import (
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/gpath"
	"regexp"
	"strings"
	"sync"
)

var (
	// data stores all settings with normalized keys (uppercase, only alphanumeric and underscores)
	data = map[string]any{}
	// mu protects concurrent access to data and callbacks
	mu sync.RWMutex
	// ConfigPath is the default path to the YAML configuration file
	ConfigPath = "./config.yml"
	// reloadCallbacks are functions called when settings are reloaded
	reloadCallbacks []func()
	// changeCallbacks are functions called when a specific setting changes
	// Supports exact matches and wildcard patterns (e.g., "DATABASE.*" or "*")
	changeCallbacks = []callbackEntry{}
)

// ChangeCallback is called when a setting value changes
// Note: The callback doesn't receive parameters about what changed.
// If you need to know which key changed, check the current value in your callback.
type ChangeCallback func()

// callbackEntry stores a callback with its pattern
type callbackEntry struct {
	pattern  string         // Original pattern (may contain wildcards)
	isExact  bool           // True if pattern has no wildcards
	regex    *regexp.Regexp // Compiled regex for wildcard patterns
	callback ChangeCallback
}

// Get retrieves a setting value by key with optional default value.
// Keys are case-insensitive and support dot notation (e.g., "DATABASE.HOST").
// Returns a generic.Value that can be converted to various types (.String(), .Int(), .Bool(), etc.)
//
// Example:
//
//	host := settings.Get("DATABASE.HOST", "localhost").String()
//	port := settings.Get("DATABASE.PORT", 3306).Int()
//	timeout := settings.Get("HTTP.TIMEOUT", "30s").Duration()
func Get(key string, defaultValue ...any) generic.Value {
	normalizedKey := normalizeKey(key)

	mu.RLock()
	v, ok := data[normalizedKey]
	mu.RUnlock()

	if ok {
		return generic.Parse(v)
	}
	if len(defaultValue) > 0 {
		return generic.Parse(defaultValue[0])
	}
	return generic.Parse("")
}

// Has checks if a setting key exists and returns its value.
// Returns (exists bool, value generic.Value).
func Has(key string) (bool, generic.Value) {
	normalizedKey := normalizeKey(key)
	mu.RLock()
	v, exists := data[normalizedKey]
	mu.RUnlock()
	return exists, generic.Parse(v)
}

// All returns a copy of all settings as a map.
// The returned map is a snapshot and safe to modify.
func All() map[string]any {
	mu.RLock()
	defer mu.RUnlock()

	result := make(map[string]any, len(data))
	for k, v := range data {
		result[k] = v
	}
	return result
}

// Set updates a setting value and triggers change callbacks.
// The key is normalized (case-insensitive, alphanumeric + underscores only).
// If database settings are enabled, the value is also persisted to the database.
//
// Example:
//
//	settings.Set("APP.NAME", "MyApp")
//	settings.Set("DATABASE.PORT", 3306)
func Set(key string, value any) {
	normalizedKey := normalizeKey(key)

	mu.Lock()
	oldValue, exists := data[normalizedKey]
	data[normalizedKey] = value
	mu.Unlock()

	// Persist to database if enabled
	if db.IsEnabled() {
		_ = saveSingleSetting(normalizedKey, value)
	}

	// Trigger change callbacks
	if exists && oldValue != value {
		triggerChangeCallbacks(normalizedKey, oldValue, value)
	} else if !exists {
		triggerChangeCallbacks(normalizedKey, nil, value)
	}
}

// SetMulti updates multiple settings at once.
// More efficient than calling Set() multiple times as it only locks once.
// If database settings are enabled, values are also persisted to the database.
func SetMulti(in map[string]any) {
	mu.Lock()
	changes := make(map[string][2]any) // key -> [oldValue, newValue]

	for key, value := range in {
		normalizedKey := normalizeKey(key)
		if oldValue, exists := data[normalizedKey]; exists && oldValue != value {
			changes[normalizedKey] = [2]any{oldValue, value}
		} else if !exists {
			changes[normalizedKey] = [2]any{nil, value}
		}
		data[normalizedKey] = value
	}
	mu.Unlock()

	// Persist to database if enabled
	if db.IsEnabled() {
		normalizedMap := make(map[string]any, len(in))
		for key, value := range in {
			normalizedMap[normalizeKey(key)] = value
		}
		db.UseModel(Setting{}, SettingDomain{})
		_ = saveDatabaseSettings(normalizedMap)

	}

	// Trigger change callbacks outside lock
	for key, change := range changes {
		triggerChangeCallbacks(key, change[0], change[1])
	}
}

// Init initializes the settings system by loading all configuration sources.
// Call this once at application startup.
func Init(params ...string) error {
	if args.Get("-c") != "" {
		ConfigPath = args.Get("-c")
	}

	return Reload()
}

// Reload reloads all settings from all sources and triggers reload callbacks.
// This is useful for hot-reloading configuration without restarting the application.
//
// Loading order (lowest to highest priority):
//  1. Environment variables
//  2. Database settings (if enabled)
//  3. YAML configuration file
//  4. Command-line arguments
func Reload() error {
	mu.Lock()
	data = map[string]any{}
	mu.Unlock()

	// 1-load environment variables
	LoadEnvVars()

	// 2-load database settings (if enabled)
	if db.IsEnabled() {
		err := LoadDatabaseSettings()
		if err != nil {
			return err
		}
	}

	// 3- load yml
	if gpath.IsFileExist(ConfigPath) {
		err := LoadYAMLSettings(ConfigPath)
		if err != nil {
			return err
		}
	}

	// 4- override args
	LoadOSArgs()

	// Trigger reload callbacks
	triggerReloadCallbacks()

	return nil
}

// setData is an internal helper to set a value without triggering callbacks.
// Used during loading from various sources.
func setData(key string, v any) {
	normalizedKey := normalizeKey(key)
	mu.Lock()
	data[normalizedKey] = v
	mu.Unlock()
}

var normalizeKeyRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)

// normalizeKey converts a key to uppercase and replaces non-alphanumeric characters with underscores.
// This ensures case-insensitive and consistent key lookups.
// Examples: "database.host" -> "DATABASE_HOST", "HTTP-Port" -> "HTTP_PORT"
func normalizeKey(arg string) string {
	arg = strings.ToUpper(arg)
	arg = normalizeKeyRegex.ReplaceAllString(arg, "_")
	return arg
}

// OnReload registers a callback to be called when settings are reloaded.
// Useful for reinitializing components when configuration changes.
//
// Example:
//
//	settings.OnReload(func() {
//	    log.Info("Configuration reloaded, reconnecting to database...")
//	    db.Reconnect()
//	})
func OnReload(callback func()) {
	mu.Lock()
	reloadCallbacks = append(reloadCallbacks, callback)
	mu.Unlock()
}

// OnChange registers a callback to be called when a setting changes.
// Supports exact matches and wildcard patterns:
//   - Exact: "DATABASE.HOST" - matches only DATABASE.HOST
//   - Prefix wildcard: "DATABASE.*" - matches DATABASE.HOST, DATABASE.PORT, etc.
//   - Global wildcard: "*" - matches all settings
//
// Note: The callback doesn't receive information about what changed.
// Use this for triggering reconnections or reloads when any matching setting changes.
//
// Examples:
//
//	// Watch specific setting and reconnect
//	settings.OnChange("DATABASE.HOST", func() {
//	    db.Reconnect()
//	})
//
//	// Watch all database settings
//	settings.OnChange("DATABASE.*", func() {
//	    db.Reconnect()
//	})
//
//	// Watch all settings
//	settings.OnChange("*", func() {
//	    app.ReloadConfig()
//	})
func OnChange(key string, callback ChangeCallback) {
	// Preserve wildcards before normalization
	hasWildcard := strings.Contains(key, "*")

	// Normalize the key but preserve wildcards temporarily
	normalizedKey := strings.ToUpper(key)

	entry := callbackEntry{
		pattern:  normalizedKey,
		callback: callback,
	}

	// Check if pattern contains wildcards
	if !hasWildcard {
		// Exact match - normalize fully and no regex needed
		entry.pattern = normalizeKey(key)
		entry.isExact = true
	} else {
		// For wildcard patterns:
		// 1. First normalize parts (convert to uppercase, replace dots with underscores)
		// 2. Then escape for regex but preserve *

		// Normalize: uppercase and replace dots/dashes with underscores, but keep *
		parts := strings.Split(normalizedKey, "*")
		for i, part := range parts {
			parts[i] = normalizeKeyRegex.ReplaceAllString(part, "_")
		}
		normalizedPattern := strings.Join(parts, "*")
		entry.pattern = normalizedPattern

		// Convert wildcard pattern to regex
		// Escape special regex characters except *
		pattern := regexp.QuoteMeta(normalizedPattern)
		// Replace escaped \* with .*
		pattern = strings.ReplaceAll(pattern, `\*`, `.*`)
		// Anchor pattern to match full key
		pattern = "^" + pattern + "$"

		var err error
		entry.regex, err = regexp.Compile(pattern)
		if err != nil {
			// If regex compilation fails, treat as exact match
			entry.isExact = true
			entry.pattern = strings.ReplaceAll(normalizedPattern, "*", "")
		}
	}

	mu.Lock()
	changeCallbacks = append(changeCallbacks, entry)
	mu.Unlock()
	callback()
}

// triggerReloadCallbacks calls all registered reload callbacks
func triggerReloadCallbacks() {
	mu.RLock()
	callbacks := make([]func(), len(reloadCallbacks))
	copy(callbacks, reloadCallbacks)
	mu.RUnlock()

	for _, callback := range callbacks {
		callback()
	}
}

// triggerChangeCallbacks calls all registered change callbacks that match the key
// Supports exact matches and wildcard patterns
func triggerChangeCallbacks(key string, oldValue, newValue any) {
	mu.RLock()
	// Collect matching callbacks
	var matchingCallbacks []ChangeCallback
	for _, entry := range changeCallbacks {
		if entry.isExact {
			// Exact match
			if entry.pattern == key {
				matchingCallbacks = append(matchingCallbacks, entry.callback)
			}
		} else {
			// Wildcard pattern match
			if entry.regex != nil && entry.regex.MatchString(key) {
				matchingCallbacks = append(matchingCallbacks, entry.callback)
			}
		}
	}
	mu.RUnlock()

	// Execute callbacks outside the lock
	for _, callback := range matchingCallbacks {
		callback()
	}
}

// SaveToYAML saves the current settings to a YAML file.
// This converts the flattened settings back to a nested structure.
//
// Example:
//
//	settings.Set("DATABASE.HOST", "localhost")
//	settings.Set("DATABASE.PORT", 3306)
//	settings.SaveToYAML("./config.yml")
//
// Results in:
//
//	database:
//	  host: localhost
//	  port: 3306
func SaveToYAML(filename string) error {
	mu.RLock()
	dataCopy := make(map[string]any, len(data))
	for k, v := range data {
		dataCopy[k] = v
	}
	mu.RUnlock()

	return saveYAMLSettings(filename, dataCopy)
}

// SaveToDB saves settings back to the database.
// Only works if database settings are enabled.
// This updates existing settings or creates new ones.
//
// Note: This saves all current settings to the database.
// Settings are stored in the default domain (domain_id = 1) unless they already exist.
func SaveToDB() error {
	if !db.IsEnabled() {
		return nil
	}

	mu.RLock()
	dataCopy := make(map[string]any, len(data))
	for k, v := range data {
		dataCopy[k] = v
	}
	mu.RUnlock()

	return saveDatabaseSettings(dataCopy)
}

// Register registers setting domains and settings definitions.
// This is typically used by modules to declare their configuration structure.
// Accepts SettingDomain and Setting structs in any order.
//
// Example:
//
//	settings.Register(
//	    settings.SettingDomain{Domain: "CACHE", Title: "Cache Settings"},
//	    settings.Setting{Domain: "CACHE", Name: "TTL", Value: "3600"},
//	)
//
// Note: Currently this function stores the definitions but does not persist them to the database.
// Future implementation will create the domains and settings in the database if enabled.
func Register(items ...any) error {
	// TODO: Implement database persistence of registered settings
	// For now, this is a no-op to maintain compatibility
	return nil
}
