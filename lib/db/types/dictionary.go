package types

import (
	"bytes"
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/getevo/json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"sync"
)

// KeyValue represents a key-value pair with generic types
type KeyValue[K comparable, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value"`
}

// Dictionary represents an array of KeyValue pairs with thread-safety
type Dictionary[K comparable, V any] struct {
	mu    sync.RWMutex
	items []KeyValue[K, V]
}

// Add adds a new key-value pair to the dictionary (Thread-Safe)
func (dict *Dictionary[K, V]) Add(key K, value V) {
	dict.mu.Lock()
	defer dict.mu.Unlock()
	dict.items = append(dict.items, KeyValue[K, V]{Key: key, Value: value})
}

// Delete removes a key-value pair by key (Thread-Safe)
func (dict *Dictionary[K, V]) Delete(key K) {
	dict.mu.Lock()
	defer dict.mu.Unlock()
	for i, kv := range dict.items {
		if kv.Key == key {
			dict.items = append(dict.items[:i], dict.items[i+1:]...)
			return
		}
	}
}

// Replace replaces a key-value pair at a specific index (Thread-Safe)
func (dict *Dictionary[K, V]) Replace(index int, key K, value V) bool {
	dict.mu.Lock()
	defer dict.mu.Unlock()
	if index >= 0 && index < len(dict.items) {
		dict.items[index] = KeyValue[K, V]{Key: key, Value: value}
		return true
	}
	return false
}

// FindByKey finds a key and returns its index and pointer to the KeyValue (Thread-Safe)
func (dict *Dictionary[K, V]) FindByKey(key K) (int, *KeyValue[K, V]) {
	dict.mu.RLock()
	defer dict.mu.RUnlock()
	for i, kv := range dict.items {
		if kv.Key == key {
			return i, &dict.items[i]
		}
	}
	return -1, nil
}

// FindByValue finds a value and returns its index and pointer to the KeyValue (Thread-Safe)
func (dict *Dictionary[K, V]) FindByValue(value V) (int, *KeyValue[K, V]) {
	dict.mu.RLock()
	defer dict.mu.RUnlock()
	for i, kv := range dict.items {
		if reflect.DeepEqual(kv.Value, value) {
			return i, &dict.items[i]
		}
	}
	return -1, nil
}

// FindByValueFunc finds a value and returns its index and pointer to the KeyValue (Thread-Safe)
func (dict *Dictionary[K, V]) FindByValueFunc(value V, equal func(a, b V) bool) (int, *KeyValue[K, V]) {
	dict.mu.RLock()
	defer dict.mu.RUnlock()
	for i, kv := range dict.items {
		if equal(kv.Value, value) {
			return i, &dict.items[i]
		}
	}
	return -1, nil
}

// String displays all key-value pairs in the dictionary (for debugging)
func (dict *Dictionary[K, V]) String() string {
	dict.mu.RLock()
	defer dict.mu.RUnlock()
	var sb strings.Builder
	for i, kv := range dict.items {
		sb.WriteString(fmt.Sprintf("[%d] Key: %v, Value: %v\n", i, kv.Key, kv.Value))
	}
	return sb.String()
}

// JSON serializes the dictionary into JSON format (Thread-Safe)
func (dict *Dictionary[K, V]) JSON() ([]byte, error) {
	dict.mu.RLock()
	defer dict.mu.RUnlock()
	if dict == nil {
		return []byte("null"), nil
	}
	if len(dict.items) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(dict.items)
}

// Value returns JSON value, implements driver.Valuer interface
func (dict *Dictionary[K, V]) Value() (driver.Value, error) {
	if dict == nil {
		return nil, nil
	}
	ba, err := dict.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return string(ba), nil
}

// Scan scans value into Dictionary, implements sql.Scanner interface
func (dict *Dictionary[K, V]) Scan(val any) error {
	dict.mu.Lock()
	defer dict.mu.Unlock()

	if val == nil {
		*dict = Dictionary[K, V]{}
		return nil
	}

	var ba []byte
	switch v := val.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return errors.New(fmt.Sprintf("Failed to unmarshal JSONB value: %v", val))
	}

	var tempItems []KeyValue[K, V]
	rd := bytes.NewReader(ba)
	decoder := json.NewDecoder(rd)
	decoder.UseNumber()
	err := decoder.Decode(&tempItems)
	if err != nil {
		return err
	}
	dict.items = tempItems
	return nil
}

// MarshalJSON serializes the dictionary to JSON (Thread-Safe)
func (dict *Dictionary[K, V]) MarshalJSON() ([]byte, error) {
	dict.mu.RLock()
	defer dict.mu.RUnlock()
	if dict == nil {
		return []byte("null"), nil
	}
	if len(dict.items) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(dict.items)
}

// UnmarshalJSON deserializes JSON into the dictionary (Thread-Safe)
func (dict *Dictionary[K, V]) UnmarshalJSON(b []byte) error {
	if b == nil || len(b) == 0 {
		return errors.New("invalid input: nil or empty JSON")
	}
	dict.mu.Lock()
	defer dict.mu.Unlock()
	return json.Unmarshal(b, &dict.items)
}

// GormDataType specifies the GORM data type for Dictionary
func (*Dictionary[K, V]) GormDataType() string {
	return "text"
}

// GormDBDataType specifies the database data type for Dictionary
func (*Dictionary[K, V]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "TEXT"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}
	return ""
}

// GormValue serializes the dictionary value for GORM
func (dict *Dictionary[K, V]) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, err := dict.MarshalJSON()
	if err != nil {
		return gorm.Expr("?", "null")
	}
	switch db.Dialector.Name() {
	case "mysql":
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}
	return gorm.Expr("?", string(data))
}
