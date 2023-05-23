# Reflections Package

The `reflections` package provides functionality for working with reflection in Go.

## Functions

### GetField

```go
func GetField(obj interface{}, name string) (interface{}, error)
```
Description: Returns the value of the provided object field.


### GetFieldKind
```go
func GetFieldKind(obj interface{}, name string) (reflect.Kind, error)
```
Description: Returns the kind of the provided object field.


### GetFieldType
```go
func GetFieldType(obj interface{}, name string) (string, error)
```
Description: Returns the type of the provided object field.

### GetFieldTag
```go
func GetFieldTag(obj interface{}, fieldName, tagKey string) (string, error)
```
Description: Returns the tag value of the provided object field.

### GetFieldNameByTagValue
```go
func GetFieldNameByTagValue(obj interface{}, tagKey, tagValue string) (string, error)
```
Description: Looks up a field with a matching tag value in the provided object.

### SetField
```go
func SetField(obj interface{}, name string, value interface{}) error
```
Description: Sets the value of the provided object field.

### HasField
```go
func HasField(obj interface{}, name string) (bool, error)
```
Description: Checks if the provided object has a field with the given name.


### Fields
```go
func Fields(obj interface{}) ([]string, error)
```
Description: Returns the names of the fields in the provided object.

### FieldsDeep
```go
func FieldsDeep(obj interface{}) ([]string, error)
```
Description: Returns the "flattened" names of the fields in the provided object, including fields from anonymous inner structs.

### Items
```go
func Items(obj interface{}) (map[string]interface{}, error)
```
Description: Returns the field:value pairs of the provided object as a map.

### ItemsDeep
```go
func ItemsDeep(obj interface{}) (map[string]interface{}, error)
```
Description: Returns the "flattened" field:value pairs of the provided object as a map, including fields from anonymous inner structs.

### Tags
```go
func Tags(obj interface{}, key string) (map[string]string, error)
```
Description: Returns the tags of the fields in the provided object as a map.

### TagsDeep
```go
func TagsDeep(obj interface{}, key string) (map[string]string, error)
```
Description: Returns the "flattened" tags of the fields in the provided object as a map, including fields from anonymous inner structs.

