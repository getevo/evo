package query

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/structtag"
	"github.com/getevo/evo"
	"github.com/getevo/evo/errors"
	"github.com/getevo/evo/lib/T"
	"github.com/getevo/evo/lib/constant"
	"github.com/getevo/evo/lib/sanitize"
	"github.com/getevo/evo/lib/validate"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

type Controller struct{}
type Filter struct {
	Object  interface{}
	Slug    string
	Allowed map[string]string
	filter  string
	values  []interface{}
	OnRow   interface{}
	MaxRows int
}
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var objects Map

func (f *Filter) SetFilter(where string, values ...interface{}) *Filter {
	if f.values == nil {
		f.values = []interface{}{}
	}
	f.values = append(f.values, values)
	f.filter += " AND (" + where + ")"
	return f
}

func (c Controller) Register(v Filter) {
	objects.Set(v.Slug, v)
	evo.Post("/filter/:slug/:id", c.Route)
}

func (Controller) Route(r *evo.Request) {
	obj := r.Params("slug")
	filter := objects.Get(obj)
	if filter != nil {
		v := reflect.ValueOf(filter.Object)
		el := v.Elem()
		t := v.Type()
		multi := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)
		pointer := reflect.New(multi.Type())
		pointer.Elem().Set(multi)
		data := pointer.Interface()

		scope := db.NewScope(el.Interface())

		var post map[string]interface{}
		err := r.BodyParser(&post)
		if err != nil {
			r.WriteResponse(constant.ERROR_JSON_PARSE)
			return
		}
		sanitize.Generic(&post)
		cond := "true" + filter.filter
		values := filter.values
		var limit int
		var offset int
		var order string
		for k, v := range post {

			if k == "offset" {
				offset = T.Must(v).Int()
				continue
			}
			if k == "limit" {
				limit = T.Must(v).Int()
				continue
			}

			if k == "order" {
				order = sanitize.String(fmt.Sprint(v))
				continue
			}
			valueType := reflect.ValueOf(v).Type().String()
			slice := strings.Split(k, "_")
			if len(slice) > 0 {
				switch slice[len(slice)-1] {
				case "like":
					key := strings.Join(slice[0:len(slice)-1], "_")
					if err := isValid(scope, filter, key, v); err != nil {
						throw(r, key, "value is not valid")
						return
					}
					cond += " AND (" + key + " LIKE ?)"
					values = append(values, "%"+fmt.Sprint(v)+"%")
					v = fmt.Sprint(v)
					break
				case "gte":
					key := strings.Join(slice[0:len(slice)-1], "_")
					if err := isValid(scope, filter, key, v); err != nil {
						throw(r, key, "value is not valid")
						return
					}
					cond += " AND (" + key + " >= ?)"
					values = append(values, v)
					break
				case "gt":
					key := strings.Join(slice[0:len(slice)-1], "_")
					if err := isValid(scope, filter, key, v); err != nil {
						throw(r, key, "value is not valid")
						return
					}
					cond += " AND (" + key + " > ?)"
					values = append(values, v)
					break
				case "lte":
					key := strings.Join(slice[0:len(slice)-1], "_")
					if err := isValid(scope, filter, key, v); err != nil {
						throw(r, key, "value is not valid")
						return
					}
					cond += " AND (" + key + " <= ?)"
					values = append(values, v)
					break
				case "lt":
					key := strings.Join(slice[0:len(slice)-1], "_")
					if err := isValid(scope, filter, key, v); err != nil {
						throw(r, key, "value is not valid")
						return
					}
					cond += " AND (" + key + " < ?)"
					values = append(values, v)
					break
				case "in":
					if len(slice) > 1 {
						if slice[len(slice)-2] == "not" {
							key := strings.Join(slice[0:len(slice)-2], "_")
							if valueType != "[]interface {}" {
								throw(r, key, "required array")
								return
							}
							if err := isValid(scope, filter, key, v); err != nil {
								throw(r, key, "value is not valid")
								return
							}
							cond += " AND (" + key + " NOT IN (?))"
							values = append(values, v)
						} else {
							key := strings.Join(slice[0:len(slice)-1], "_")
							if valueType != "[]interface {}" {
								throw(r, key, "required array")
								return
							}
							if err := isValid(scope, filter, key, v); err != nil {
								throw(r, key, "value is not valid")
								return
							}
							cond += " AND (" + key + " IN (?))"
							values = append(values, v)
						}

					}
					break
				case "between":
					key := strings.Join(slice[0:len(slice)-1], "_")
					if valueType != "[]interface {}" || len(v.([]interface{})) != 2 {
						throw(r, key, "required two positional args as array")
						return
					}
					if err := isValid(scope, filter, key, v.([]interface{})[0]); err != nil {
						throw(r, key, "value is not valid")
						return
					}

					if err := isValid(scope, filter, key, v.([]interface{})[1]); err != nil {
						throw(r, key, "value is not valid")
						return
					}

					cond += " AND (" + key + " BETWEEN ? AND ?)"
					values = append(values, v.([]interface{})[0], v.([]interface{})[1])
					break
				default:
					if err := isValid(scope, filter, k, v); err != nil {
						throw(r, k, "value is not valid")
						return
					}
					cond += " AND (" + k + " = ?)"
					values = append(values, v)
					break
				}
			}
		}
		if limit < 1 {
			limit = 10
		}
		if offset < 0 {
			offset = 0
		}

		query := db.Unscoped().Offset(offset).Limit(limit).Where(cond, values...)
		if order != "" {
			query = query.Order(order)
		}
		query.Find(data)
		r.WriteResponse(data)
		return
	}
	r.WriteResponse(constant.ERROR_OBJECT_NOT_EXIST)
}

func throw(r *evo.Request, field, err string) {
	r.WriteResponse(e.Field(field, err))
}

func isValid(scope *gorm.Scope, filter *Filter, key string, val interface{}) error {
	for _, item := range scope.Fields() {
		jsonTag := item.Tag.Get("json")
		formTag := item.Tag.Get("form")
		gormTag := item.Tag.Get("gorm")
		if gormTag != "-" {
			if validator, ok := filter.Allowed[key]; ok && ((jsonTag != "-" && key == jsonTag) || (formTag != "-" && key == formTag)) {
				tags, err := structtag.Parse(validator)
				if err != nil {
					return err
				}
				if tag, err := tags.Get("validate"); err == nil {
					if err = validate.ValidateVariable(key, tag.Value(), val); err != nil {
						spew.Dump(err)
						return err
					}
				} else {
					return err
				}

			}
		}
	}
	return nil
}
