package generic

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"
)

const (
	Invalid reflect.Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)

// Parse parse input
//
//	@param i
//	@return Value
func Parse(i any) Value {
	return Value{
		Input: i,
	}
}

// Type lib structure to keep input and its type
type Type struct {
	input any
	iType reflect.Type
}

// Value wraps over interface
type Value struct {
	Input any
}

// IsNil returns if the value is nil
//
//	@receiver v
//	@return bool
func (v Value) IsNil() bool {
	return !reflect.ValueOf(v.Input).IsValid()
}

func (v Value) direct() any {
	ref := reflect.ValueOf(v.Input)
	if v.Input == nil {
		return ""
	}
	if !ref.IsValid() {
		return nil
	}
	if ref.Type().Kind() == reflect.Ptr {
		return ref.Elem().Interface()
	}
	return v.Input
}

// ParseJSON parse json value into struct
//
//	@receiver v
//	@param in
//	@return error
func (v Value) ParseJSON(in any) error {
	var value = v.direct()
	return json.Unmarshal([]byte(fmt.Sprint(value)), in)
}

// String return value as string
//
//	@receiver v
//	@return string
func (v Value) String() string {
	var value = v.direct()
	return fmt.Sprint(value)
}

// Int return value as integer
//
//	@receiver v
//	@return int
func (v Value) Int() int {
	if float, ok := v.Input.(float64); ok {
		return int(float)
	} else if float, ok := v.Input.(float32); ok {
		return int(float)
	}
	i, _ := strconv.Atoi(v.String())
	return i
}

func (v Value) Int8() int8 {
	i, _ := strconv.ParseInt(v.String(), 10, 8)
	return int8(i)
}

func (v Value) Int16() int16 {
	i, _ := strconv.ParseInt(v.String(), 10, 16)
	return int16(i)
}

func (v Value) Int32() int32 {
	i, _ := strconv.ParseInt(v.String(), 10, 32)
	return int32(i)
}

func (v Value) Uint() uint {
	i, _ := strconv.ParseUint(v.String(), 10, 32)
	return uint(i)
}

func (v Value) Uint8() uint8 {
	i, _ := strconv.ParseUint(v.String(), 10, 8)
	return uint8(i)
}

func (v Value) Uint16() uint16 {
	i, _ := strconv.ParseUint(v.String(), 10, 16)
	return uint16(i)
}

func (v Value) Uint32() uint32 {
	i, _ := strconv.ParseUint(v.String(), 10, 32)
	return uint32(i)
}

func (v Value) Float32() float32 {
	i, _ := strconv.ParseFloat(v.String(), 32)
	return float32(i)
}

func (v Value) Float64() float64 {
	return v.Float()
}

// Uint64 return value as uint64
//
//	@receiver v
//	@return uint64
func (v Value) Uint64() uint64 {
	if float, ok := v.Input.(float64); ok {
		return uint64(float)
	} else if float, ok := v.Input.(float32); ok {
		return uint64(float)
	}
	i, _ := strconv.ParseUint(v.String(), 0, 64)
	return i
}

// Int64 return value as int64
//
//	@receiver v
//	@return int64
func (v Value) Int64() int64 {
	if float, ok := v.Input.(float64); ok {
		return int64(float)
	} else if float, ok := v.Input.(float32); ok {
		return int64(float)
	}
	i, _ := strconv.ParseInt(v.String(), 0, 64)
	return i
}

// Float return value as float64
//
//	@receiver v
//	@return float64
func (v Value) Float() float64 {
	if float, ok := v.Input.(float64); ok {
		return float
	} else if float, ok := v.Input.(float32); ok {
		return float64(float)
	}
	i, _ := strconv.ParseFloat(v.String(), 64)
	return i
}

// Bool return value as bool
//
//	@receiver v
//	@return bool
func (v Value) Bool() bool {
	var s = strings.ToLower(v.String())
	if len(s) > 0 {
		if s == "1" || s == "true" || s == "yes" {
			return true
		}
	}
	return false
}

// Time return value as time.Time
//
//	@receiver v
//	@return time.Time
//	@return error
func (v Value) Time() (time.Time, error) {
	return dateparse.ParseAny(v.String())
}

// Duration return value as time.Duration
//
//	@receiver v
//	@return time.Duration
//	@return error
func (v Value) Duration() (time.Duration, error) {
	return time.ParseDuration(v.String())
}

// ToString cast anything to string
//
//	@param v
//	@return string
func ToString(v any) string {
	ref := reflect.ValueOf(v)
	if !ref.IsValid() {
		return ""
	}
	if ref.Type().Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	switch ref.Kind() {
	case reflect.String:
		return ref.Interface().(string)
	case reflect.Struct, reflect.Slice:
		if v, ok := ref.Interface().([]byte); ok {
			return string(v)
		}
		if v, ok := ref.Interface().(Value); ok {
			return v.String()
		}
		b, _ := json.Marshal(ref.Interface())
		return string(b)
	default:
		return fmt.Sprint(ref.Interface())
	}
}

// TypeOf return type of input
//
//	@param input
//	@return *Type
func TypeOf(input any) *Type {
	var el = Type{
		input: input,
		iType: reflect.TypeOf(input),
	}
	return &el
}

// Is checks if input and given type are equal
//
//	@receiver t
//	@param input
//	@return bool
func (t *Type) Is(input any) bool {
	switch v := input.(type) {
	case reflect.Kind:
		return v == t.iType.Kind()
	case string:
		return t.iType.String() == v
	}
	return TypeOf(input).iType.String() == t.iType.String()
}

// Indirect get type of object considering if it is pointer
//
//	@receiver t
//	@return *Type
func (t *Type) Indirect() *Type {
	if t.Is(reflect.Ptr) {
		var el = Type{
			input: reflect.ValueOf(t.input).Elem().Interface(),
		}
		el.iType = reflect.TypeOf(el.input)
		return &el
	}
	return t
}

var sizeRegex = regexp.MustCompile(`(?m)^(\d+)\s*([kmgte]{0,1}b){0,1}$`)

func (v Value) SizeInBytes() uint64 {
	var s = strings.ToLower(strings.TrimSpace(fmt.Sprint(v.Input)))
	var match = sizeRegex.FindAllStringSubmatch(s, 1)
	if len(match) == 1 && len(match[0]) == 3 {
		var base, _ = strconv.ParseUint(match[0][1], 10, 64)
		switch match[0][2] {
		case "kb":
			base *= 1024
		case "mb":
			base *= 1024 * 1024
		case "gb":
			base *= 1024 * 1024 * 1024
		case "tb":
			base *= 1024 * 1024 * 1024 * 1024
		case "eb":
			base *= 1024 * 1024 * 1024 * 1024 * 1024
		}

		return base
	}
	return 0
}

func (v Value) ByteCount() string {
	var b, _ = strconv.ParseUint(strings.TrimSpace(fmt.Sprint(v.Input)), 10, 64)
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func (v *Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Input)
}

func (v *Value) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.Input)
}

func (v *Value) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(v.Input)
}

func (v *Value) UnmarshalYAML(data []byte) error {
	return yaml.Unmarshal(data, &v.Input)
}

func (v *Value) Scan(value any) error {
	switch cast := value.(type) {
	case string:
		v.Input = cast
	case []byte:
		v.Input = string(cast)
	default:
		v.Input = cast
	}
	return nil
}

// Value return drive.Value value, implement driver.Valuer interface of gorm
func (v Value) Value() (driver.Value, error) {
	return ToString(v.Input), nil
}

func (v Value) Is(s string) bool {
	var typ = reflect.TypeOf(v.Input).String()
	return typ == s
}

func (v Value) SameAs(s any) bool {
	return reflect.TypeOf(v.Input) == reflect.TypeOf(s)
}

func (v Value) Prop(property string) Value {
	ref := v.Indirect()
	if ref.Kind() == reflect.Struct {
		return Parse(ref.FieldByName(property).Interface())
	}
	if ref.Kind() == reflect.Map {
		return Parse(ref.MapIndex(reflect.ValueOf(property)).Interface())
	}
	return Parse(nil)
}

func (v Value) PropByTag(tag string) Value {
	ref := v.Indirect()
	if ref.Kind() == reflect.Struct {
		var typ = reflect.TypeOf(ref.Interface())
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.Tag.Get(tag) == tag {
				return Parse(ref.Field(i).Interface())
			}
		}
	}

	return Parse(nil)
}

func (v Value) SetProp(property string, value any) error {
	ref := v.Indirect()
	if ref.Kind() == reflect.Invalid {
		return nil
	}

	if ref.Kind() == reflect.Struct {
		if ref.Type() == reflect.TypeOf(time.Time{}) {
			t, err := Parse(value).Time()
			if err != nil {
				return err
			}
			ref.Set(reflect.ValueOf(t))
			return nil
		}
		//var x = ref.FieldByName(property).Interface()

		var field = ref.FieldByName(property)
		if field.Kind() == reflect.Struct {
			return Value{Input: field.Addr().Interface()}.SetProp(property, value)
		}
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}
		var err = Parse(value).Cast(field)
		if err != nil {
			return err
		}
		//ref.FieldByName(property).Set(reflect.ValueOf(x).Convert(field.Type()))

		return nil
	}
	if ref.Kind() == reflect.Map {
		typ := v.IndirectType()
		var nv = reflect.New(typ.Elem())
		var err = Parse(value).Cast(nv)
		if err != nil {
			log.Error(err)
		}
		ref.SetMapIndex(reflect.ValueOf(property), nv.Elem())
		return nil
	}
	return fmt.Errorf("value is not struct or map")
}

func (v Value) Cast(dst any) error {
	var ref reflect.Value
	if v, ok := dst.(reflect.Value); ok {
		ref = v
	} else {
		ref = reflect.ValueOf(dst)
	}
	for ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	var kind = ref.Kind()
	if kind == reflect.Struct {
		x := Parse(dst)
		var vRef = v.Indirect()

		for _, prop := range v.Props() {

			if vRef.Kind() == reflect.Struct {
				if x.HasProp(prop.Name) {
					err := x.SetProp(x.GetName(prop.Name), vRef.FieldByName(prop.Name).Interface())
					if err != nil {
						return err
					}
				}
			} else if vRef.Kind() == reflect.Map {

				if idx := vRef.MapIndex(reflect.ValueOf(prop.Name)); idx.IsValid() {
					err := x.SetProp(x.GetName(prop.Name), idx.Interface())
					if err != nil {
						return err
					}
					continue
				}
				if idx := vRef.MapIndex(reflect.ValueOf(strcase.ToSnake(prop.Name))); idx.IsValid() {
					err := x.SetProp(x.GetName(prop.Name), idx.Interface())
					if err != nil {
						return err
					}
				}

			} else {
				return fmt.Errorf("cant cast %s to %s", ref.Kind().String(), vRef.Kind().String())
			}

		}

		return nil
	}

	if kind == reflect.Map {
		x := Parse(dst)

		for _, prop := range v.Props() {
			err := x.SetProp(prop.Name, v.Prop(prop.Name).Input)
			if err != nil {
				return err
			}
		}
	}

	var x any
	switch kind {
	case reflect.Int:
		if sizeRegex.MatchString(v.String()) {
			x = int(v.SizeInBytes())
		} else {
			x = v.Int()
		}

	case reflect.Int8:
		if sizeRegex.MatchString(v.String()) {
			x = int8(v.SizeInBytes())
		} else {
			x = v.Int8()
		}
	case reflect.Int16:
		if sizeRegex.MatchString(v.String()) {
			x = int16(v.SizeInBytes())
		} else {
			x = v.Int16()
		}
	case reflect.Int32:
		if sizeRegex.MatchString(v.String()) {
			x = int32(v.SizeInBytes())
		} else {
			x = v.Int32()
		}
	case reflect.Int64:

		if ref.Type().String() == "time.Duration" {
			var t, err = v.Duration()
			if err != nil {
				return err
			}
			x = t
		} else {
			if sizeRegex.MatchString(v.String()) {
				x = int64(v.SizeInBytes())
			} else {
				x = v.Int64()
			}
		}
	case reflect.Uint:
		if sizeRegex.MatchString(v.String()) {
			x = uint(v.SizeInBytes())
		} else {
			x = v.Uint()
		}
	case reflect.Uint8:
		if sizeRegex.MatchString(v.String()) {
			x = uint(v.SizeInBytes())
		} else {
			x = v.Uint()
		}
	case reflect.Uint16:
		if sizeRegex.MatchString(v.String()) {
			x = uint16(v.SizeInBytes())
		} else {
			x = v.Uint16()
		}
	case reflect.Uint32:
		if sizeRegex.MatchString(v.String()) {
			x = uint32(v.SizeInBytes())
		} else {
			x = v.Uint32()
		}
	case reflect.Uint64:
		if sizeRegex.MatchString(v.String()) {
			x = uint32(v.SizeInBytes())
		} else {
			x = v.Uint64()
		}
	case reflect.String:
		x = v.String()
	case reflect.Bool:
		x = v.Bool()
	case reflect.Float32:
		x = v.Float32()
	case reflect.Float64:
		x = v.Float64()
	case reflect.Func, reflect.Struct, reflect.Interface:
		return nil
	default:
		return fmt.Errorf("couldnt convert to %s %s", ref.String(), v.String())
	}
	ref.Set(reflect.ValueOf(x).Convert(ref.Type()))
	return nil
}

func (v Value) Props() []reflect.StructField {
	var typ = v.IndirectType()
	var val = v.Indirect()
	if typ.Kind() == reflect.Map {
		var ref = v.Indirect()

		var keys []reflect.StructField
		for _, item := range ref.MapKeys() {
			keys = append(keys, reflect.StructField{
				Name:    fmt.Sprint(item),
				PkgPath: "-",
			})
		}
		return keys
	}
	var fields []reflect.StructField
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Type.Kind() == reflect.Struct {
			fields = append(fields, Parse(val.Field(i).Interface()).Props()...)
		} else {
			fields = append(fields, typ.Field(i))
		}

	}
	return fields
}

func (v Value) HasProp(name string) bool {
	var typ = v.IndirectType()
	if typ.Kind() == reflect.Map {
		return v.Indirect().MapIndex(reflect.ValueOf(name)).IsValid()
	}
	for i := 0; i < typ.NumField(); i++ {
		var field = typ.Field(i)
		if field.Name == name {
			return true
		}
	}
	return false
}

func (v Value) GetName(name string) string {
	var typ = v.IndirectType()

	toMatch := strings.ReplaceAll(strings.ToLower(name), "_", "")

	for i := 0; i < typ.NumField(); i++ {
		var field = typ.Field(i)
		if strings.ReplaceAll(strings.ToLower(field.Name), "_", "") == toMatch {
			return field.Name
		}
	}
	return ""
}

func (v Value) Indirect() reflect.Value {
	return indirect(v.Input)
}

func indirect(input any) reflect.Value {
	var ref = reflect.ValueOf(input)
	for ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	return ref
}

func (v Value) IndirectType() reflect.Type {
	return indirectType(v.Input)
}

func indirectType(input any) reflect.Type {
	var ref = reflect.TypeOf(input)
	for ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	return ref
}

func (v Value) IsEmpty() bool {
	return v.Input == nil || reflect.ValueOf(v.Input).IsNil()
}

func (v Value) IsAny(s ...any) bool {
	var t = v.IndirectType()
	var kind = t.Kind()
	for _, item := range s {
		switch v := item.(type) {
		case reflect.Kind:
			if v == kind {
				return true
			}
		case string:
			if strings.ToLower(t.String()) == v {
				return true
			}
		default:
			if TypeOf(v).iType.String() == t.String() {
				return true
			}
		}
	}
	return false
}

func (v Value) New() Value {
	return Value{Input: reflect.New(v.IndirectType()).Interface()}
}

func (v Value) FieldNames() []string {
	var typ = v.IndirectType()
	var val = v.Indirect()
	var fields []string
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Type.Kind() == reflect.Struct {
			fields = append(fields, Parse(val.Field(i).Interface()).FieldNames()...)
		} else {
			fields = append(fields, typ.Field(i).Name)
		}

	}
	return fields
}
