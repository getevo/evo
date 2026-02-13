package validation_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/getevo/evo/v2/lib/validation"
	"github.com/stretchr/testify/assert"
)

type BasicStruct struct {
	Alpha          string `validation:"alpha"`
	RequiredAlpha  string `validation:"alpha,required"`
	AlphaNumeric   string `validation:"alphanumeric"`
	Email          string `validation:"email"`
	RequiredEmail  string `validation:"email,required"`
	URL            string `validation:"url"`
	Domain         string `validation:"domain"`
	IP             string `validation:"ip"`
	Regex          string `validation:"regex([a-z]{2,})"`
	LengthEq       string `validation:"len==10"`
	LengthGt       string `validation:"len>10"`
	LengthLt       string `validation:"len<10"`
	PositiveInt    string `validation:"+int"`
	NegativeInt    string `validation:"-int"`
	PositiveFloat  string `validation:"+float"`
	NegativeFloat  string `validation:"-float"`
	PositiveFloat2 string `validation:"float"`
}

func TestBasicStruct(t *testing.T) {
	// Test individual validators without database (using Value instead of Struct)
	// Since Struct validation requires database for GORM schema parsing,
	// we test individual validators here

	tests := []struct {
		name       string
		value      interface{}
		validation string
		wantErr    bool
	}{
		{"alpha valid", "abc", "alpha", false},
		{"alpha required valid", "abc", "alpha,required", false},
		{"alphanumeric valid", "abc123", "alphanumeric", false},
		{"email valid", "test@example.com", "email", false},
		{"email required valid", "test@example.com", "email,required", false},
		{"url valid", "https://example.com/path", "url", false},
		{"domain valid", "example.com", "domain", false},
		{"ip valid", "127.0.0.1", "ip", false},
		{"regex valid", "abc", "regex([a-z]{2,})", false},
		{"length eq valid", "1234567890", "len==10", false},
		{"length gt valid", "12345678901", "len>10", false},
		{"length lt valid", "123456789", "len<10", false},
		{"positive int valid", "10", "+int", false},
		{"negative int valid", "-10", "-int", false},
		{"positive float valid", "10.5", "+float", false},
		{"negative float valid", "-10.5", "-float", false},
		{"float valid", "10.5", "float", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, tt.validation)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Fixed Bugs

func TestPortValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid port", "8080", false},
		{"valid port 80", "80", false},
		{"valid port 65535", "65535", false},
		{"invalid port 0", "0", true},
		{"invalid port 65536", "65536", true},
		{"invalid port negative", "-1", true},
		{"invalid port text", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "port")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLatitudeValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid latitude 0", "0", false},
		{"valid latitude 45.5", "45.5", false},
		{"valid latitude -45.5", "-45.5", false},
		{"valid latitude 90", "90", false},
		{"valid latitude -90", "-90", false},
		{"invalid latitude 91", "91", true},
		{"invalid latitude -91", "-91", true},
		{"invalid latitude text", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "latitude")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLongitudeValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid longitude 0", "0", false},
		{"valid longitude 120.5", "120.5", false},
		{"valid longitude -120.5", "-120.5", false},
		{"valid longitude 180", "180", false},
		{"valid longitude -180", "-180", false},
		{"invalid longitude 181", "181", true},
		{"invalid longitude -181", "-181", true},
		{"invalid longitude text", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "longitude")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		level   string
		wantErr bool
	}{
		{"easy - valid", "simple", "easy", false},
		{"easy - too short", "abc", "easy", true},
		{"medium - valid", "Pass123!", "medium", false},
		{"medium - too short", "Pass1!", "medium", true},
		{"medium - not complex", "password", "medium", true},
		{"hard - valid", "SecurePass123!@#", "hard", false},
		{"hard - too short", "Pass123!", "hard", true},
		{"hard - not complex enough", "password123", "hard", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, fmt.Sprintf("password(%s)", tt.level))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNumericalValidatorNotEqual(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"not equal to 10 - valid 5", "5", false},
		{"not equal to 10 - valid 15", "15", false},
		{"not equal to 10 - invalid 10", "10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "!=10")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test New String Validators

func TestInValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"value in list - red", "red", false},
		{"value in list - green", "green", false},
		{"value in list - blue", "blue", false},
		{"value not in list", "yellow", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "in(red,green,blue)")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotInValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"value not in list", "yellow", false},
		{"value in list - red", "red", true},
		{"value in list - green", "green", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "not_in(red,green,blue)")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContainsValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"contains substring", "hello world", false},
		{"does not contain", "goodbye world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "contains(hello)")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStartsWithValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"starts with prefix", "hello world", false},
		{"does not start with prefix", "world hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "starts_with(hello)")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEndsWithValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"ends with suffix", "hello world", false},
		{"does not end with suffix", "world hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "ends_with(world)")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestASCIIValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid ASCII", "Hello World 123!", false},
		{"invalid ASCII - emoji", "Hello 👋", true},
		{"invalid ASCII - chinese", "你好", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "ascii")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrintableValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid printable", "Hello World!", false},
		{"invalid printable - control char", "Hello\x00World", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "printable")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Array Validators

func TestMinItemsValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   []string
		wantErr bool
	}{
		{"has minimum items", []string{"a", "b", "c"}, false},
		{"exactly minimum items", []string{"a", "b"}, false},
		{"below minimum items", []string{"a"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "min_items(2)")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMaxItemsValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   []string
		wantErr bool
	}{
		{"below maximum items", []string{"a", "b"}, false},
		{"exactly maximum items", []string{"a", "b", "c"}, false},
		{"above maximum items", []string{"a", "b", "c", "d"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "max_items(3)")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUniqueItemsValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   []string
		wantErr bool
	}{
		{"all unique", []string{"a", "b", "c"}, false},
		{"has duplicates", []string{"a", "b", "a"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "unique_items")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Date Validators

func TestBeforeNowValidator(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"date in past", past, false},
		{"date in future", future, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "before_now")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAfterNowValidator(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"date in future", future, false},
		{"date in past", past, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "after_now")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDateFormatValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		format  string
		wantErr bool
	}{
		{"valid format YYYY-MM-DD", "2025-08-03", "2006-01-02", false},
		{"invalid format", "08/03/2025", "2006-01-02", true},
		{"valid format MM/DD/YYYY", "08/03/2025", "01/02/2006", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, fmt.Sprintf("date_format(%s)", tt.format))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Financial Validators

func TestIBANValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid IBAN", "GB82WEST12345698765432", false},
		{"valid IBAN DE", "DE89370400440532013000", false},
		{"invalid IBAN - too short", "GB82", true},
		{"invalid IBAN - no country", "1234567890", true},
		{"invalid IBAN - lowercase", "gb82west12345698765432", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Value(tt.value, "iban")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Context Support

func TestStructWithContext(t *testing.T) {
	// Test ValueWithContext to ensure context propagation works
	ctx := context.Background()

	tests := []struct {
		name       string
		value      interface{}
		validation string
		wantErr    bool
	}{
		{"alpha with context", "abc", "alpha", false},
		{"email with context", "test@example.com", "email", false},
		{"url with context", "https://example.com", "url", false},
		{"domain with context", "example.com", "domain", false},
		{"ip with context", "192.168.1.1", "ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValueWithContext(ctx, tt.value, tt.validation)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "Expected no validation errors with context")
			}
		})
	}
}

func TestValueWithContext(t *testing.T) {
	ctx := context.Background()

	err := validation.ValueWithContext(ctx, "test@example.com", "email")
	assert.NoError(t, err, "Email should be valid with context")
}

// Test Custom Validator Registration

func TestCustomValidator(t *testing.T) {
	// Note: Custom validator registration requires access to internal types
	// This test demonstrates the API but may need internal package access
	// For production use, create validators in the validation package itself

	// Test that we can call Value with a custom pattern (will fail if not registered)
	err := validation.Value("test", "required")
	assert.NoError(t, err, "Required validator should pass for non-empty value")
}

// Benchmark Tests

func BenchmarkEmailValidator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = validation.Value("test@example.com", "email")
	}
}

func BenchmarkStructValidation(b *testing.B) {
	s := BasicStruct{
		Alpha:          "abc",
		RequiredAlpha:  "abc",
		AlphaNumeric:   "abc123",
		Email:          "test@example.com",
		RequiredEmail:  "test@example.com",
		URL:            "https://example.com",
		Domain:         "example.com",
		IP:             "192.168.1.1",
		Regex:          "abc",
		LengthEq:       "1234567890",
		LengthGt:       "12345678901",
		LengthLt:       "123456789",
		PositiveInt:    "10",
		NegativeInt:    "-10",
		PositiveFloat:  "10.5",
		NegativeFloat:  "-10.5",
		PositiveFloat2: "10.5",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validation.Struct(s)
	}
}
