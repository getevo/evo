package html

import (
	"fmt"
	"reflect"
	"strings"
)

type Attributes map[string]interface{}
type Element struct {
	Tag        string
	Body       interface{}
	Attributes Attributes
}

var ViewKey = "html"

func (attrs Attributes) Render() string {
	var res = ""
	for k, v := range attrs {
		res += " " + k + "=\"" + strings.Replace(fmt.Sprint(v), "\"", "\\\"", -1) + "\""
	}
	return res
}

func (attrs Attributes) Set(key string, value interface{}) {
	attrs[key] = value
}

func Tag(tag string, inner interface{}) *Element {
	el := Element{
		Tag:        tag,
		Body:       inner,
		Attributes: Attributes{},
	}
	return &el
}

func (el *Element) Set(key string, value interface{}) *Element {
	el.Attributes.Set(key, value)
	return el
}

func (el Element) Render() string {

	return "<" + el.Tag + " " + el.Attributes.Render() + ">" + Render(el.Body) + "</" + el.Tag + ">"
}

func Render(el interface{}) string {

	kind := reflect.TypeOf(el).Kind()
	if kind == reflect.Ptr {
		return Render(reflect.ValueOf(el).Elem().Interface())
	}
	if kind == reflect.Slice || kind == reflect.Array {
		var str string
		s := reflect.ValueOf(el)
		for i := 0; i < s.Len(); i++ {
			str += "\r\n\t" + Render(s.Index(i).Interface())
		}
		return str
	}

	if v, ok := el.(Element); ok {
		return v.Render()
	}
	if v, ok := el.(InputStruct); ok {
		return v.Render()
	}
	if v, ok := el.(string); ok {
		return v
	}
	return fmt.Sprint(el)
}

func Icon(icon string) string {
	return "<i class=\"fa fa-" + icon + "\"></i>"
}
