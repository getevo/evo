package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// InternalFunctions groups equivalent function names for default-value comparison.
// Used by both MySQL and PostgreSQL dialects when comparing default values.
var InternalFunctions = [][]string{
	{"CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP()", "current_timestamp()", "current_timestamp", "NOW()", "now()"},
	{"CURRENT_DATE", "CURRENT_DATE()", "current_date", "current_date()"},
	{"NULL", "null"},
}

// MigrationResult holds the output of a dialect's GenerateMigration call.
type MigrationResult struct {
	Queries     []string        // CREATE/ALTER statements
	Tail        []string        // FK constraints (deferred)
	TableExists map[string]bool // which tables exist (for version migrations)
}

// JoinConstraint represents a foreign key relationship between two tables.
type JoinConstraint struct {
	Table            string
	Column           string
	ReferencedTable  string
	ReferencedColumn string
}

// Column represents a model column definition used by the ColumnDefinition callback.
type Column struct {
	Name          string
	Nullable      bool
	PrimaryKey    bool
	Size          int
	Scale         int
	Precision     int
	Type          string
	Default       string
	AutoIncrement bool
	Unique        bool
	Comment       string
	Charset       string
	OnUpdate      string
	OnDelete      string
	FKOnUpdate    string
	Collate       string
	ForeignKey    string
	After         string
	FullText      bool
}

// Columns is a slice of Column with helper methods.
type Columns []Column

// Find returns a pointer to the column with the given name, or nil.
func (list Columns) Find(name string) *Column {
	for idx := range list {
		if list[idx].Name == name {
			return &list[idx]
		}
	}
	return nil
}

// Keys returns the column names.
func (list Columns) Keys() []string {
	var result []string
	for _, item := range list {
		result = append(result, item.Name)
	}
	return result
}

// Index represents a model index definition.
type Index struct {
	Name     string
	Unique   bool
	FullText bool
	Columns  Columns
}

// Indexes is a slice of Index with helper methods.
type Indexes []Index

// Find returns a pointer to the index with the given name, or nil.
func (list Indexes) Find(name string) *Index {
	for idx := range list {
		if list[idx].Name == name {
			return &list[idx]
		}
	}
	return nil
}

// --- Shared utility functions ---

// CleanEnum strips whitespace outside quotes from an enum definition.
func CleanEnum(str string) string {
	var result = ""
	var inside = false
	for _, char := range str {
		if char == '\'' || char == '"' || char == '`' {
			inside = !inside
		}
		if !inside && char == ' ' {
			continue
		}
		result += string(char)
	}
	return result
}

// TrimQuotes removes matching surrounding quotes from a string.
func TrimQuotes(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// Generate32CharHash takes a text input and returns a unique 32-character hash.
func Generate32CharHash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])[:32]
}

// SafeIndexName truncates an index name to maxLen characters if it exceeds
// the limit, appending a short hash to preserve uniqueness.
// MySQL limit: 64, PostgreSQL limit: 63.
func SafeIndexName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	h := Generate32CharHash(name)[:8]
	return name[:maxLen-9] + "_" + h
}

// --- Config map for dialect defaults ---

var (
	configMu sync.RWMutex
	configMap = map[string]string{}
)

// SetConfig stores a key-value pair in the shared config map.
func SetConfig(key, value string) {
	configMu.Lock()
	defer configMu.Unlock()
	configMap[key] = value
}

// GetConfig retrieves a value from the shared config map.
// Returns the value and true if found, or empty string and false otherwise.
func GetConfig(key string) (string, bool) {
	configMu.RLock()
	defer configMu.RUnlock()
	v, ok := configMap[key]
	return v, ok
}

// GetConfigDefault retrieves a value from the shared config map, returning
// the provided default if the key is not found.
func GetConfigDefault(key, defaultValue string) string {
	configMu.RLock()
	defer configMu.RUnlock()
	if v, ok := configMap[key]; ok {
		return v
	}
	return defaultValue
}
