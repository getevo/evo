package ptr

import "time"

// Int returns a pointer to the given int value.
func Int(v int) *int {
	return &v
}

// Int8 returns a pointer to the given int8 value.
func Int8(v int8) *int8 {
	return &v
}

// Int16 returns a pointer to the given int16 value.
func Int16(v int16) *int16 {
	return &v
}

// Int32 returns a pointer to the given int32 value.
func Int32(v int32) *int32 {
	return &v
}

// Int64 returns a pointer to the given int64 value.
func Int64(v int64) *int64 {
	return &v
}

// Uint returns a pointer to the given uint value.
func Uint(v uint) *uint {
	return &v
}

// Uint8 returns a pointer to the given uint8 value.
func Uint8(v uint8) *uint8 {
	return &v
}

// Uint16 returns a pointer to the given uint16 value.
func Uint16(v uint16) *uint16 {
	return &v
}

// Uint32 returns a pointer to the given uint32 value.
func Uint32(v uint32) *uint32 {
	return &v
}

// Uint64 returns a pointer to the given uint64 value.
func Uint64(v uint64) *uint64 {
	return &v
}

// Float32 returns a pointer to the given float32 value.
func Float32(v float32) *float32 {
	return &v
}

// Float64 returns a pointer to the given float64 value.
func Float64(v float64) *float64 {
	return &v
}

// String returns a pointer to the given string value.
func String(v string) *string {
	return &v
}

// Bool returns a pointer to the given bool value.
func Bool(v bool) *bool {
	return &v
}

// Time returns a pointer to the given time.Time value.
func Time(v time.Time) *time.Time {
	return &v
}

// Interface returns a pointer to the given interface{} value.
func Interface(v interface{}) *interface{} {
	return &v
}
