package settings

import (
	"encoding/json"
	"github.com/getevo/evo"
	"github.com/getevo/evo/errors"
	"github.com/getevo/evo/html"
	"github.com/getevo/evo/i18"
	"github.com/getevo/evo/lib"
	"github.com/getevo/evo/lib/constant"
	"github.com/getevo/evo/lib/ref"
	"github.com/getevo/evo/lib/text"
	"reflect"
	"strings"
)

type Controller struct{}
type Registry struct {
	Title     string
	Form      []html.InputStruct
	Slug      string
	Reference string
}

func (c Controller) set(s string, object interface{}) {
	ref := reflect.ValueOf(object)
	_type := strings.ToLower(ref.Elem().Type().String())

	_default := text.ToJSON(object)
	obj := Settings{
		Reference: _type,
		Title:     s,
		Data:      _default,
		Default:   _default,
		Ptr:       object,
	}
	settings.Set(_type, obj)
	if db.Where("reference = ?", _type).Take(&obj).RowsAffected == 0 {
		db.Create(&obj)
		return
	}
	json.Unmarshal([]byte(obj.Data), object)
}

func (c Controller) view(r *evo.Request) {
	if r.User.Anonymous {
		r.Flash("warning", constant.ERROR_UNAUTHORIZED.Error())
		r.Redirect("/admin/error")
		return
	}
	if !r.User.HasPerm("settings.access") {
		r.Flash("error", constant.ERROR_MUST_LOGIN.Error())
		r.Redirect("/admin/login")
		return
	}
	r.Var("heading", "Settings")

	var registries []Registry
	for _, key := range settings.Keys() {
		v := settings.Get(key).(Settings)

		registries = append(registries, Registry{
			v.Title, c.getForm(v.Ptr), text.Slugify(v.Reference), v.Reference,
		})
	}

	r.View(map[string]interface{}{
		"registries": registries,
	}, "settings.settings", "template.default")
}

func (c Controller) getForm(v interface{}) []html.InputStruct {
	ref := reflect.ValueOf(v).Elem()
	frm := []html.InputStruct{}
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Type().Field(i)
		tags := field.Tag
		if len(tags) == 0 {
			continue
		}
		input := html.Input("text", field.Name, field.Name)

		input.SetValue(ref.Field(i).Interface())

		if typ, ok := tags.Lookup("type"); ok {
			input.Type = typ
		}
		if hint, ok := tags.Lookup("hint"); ok {
			input.Hint = hint
		}
		if label, ok := tags.Lookup("label"); ok {
			input.SetLabel(label)
		}
		if col, ok := tags.Lookup("col"); ok {
			input.SetSize(lib.ParseSafeInt(col))
		}
		if min, ok := tags.Lookup("min"); ok {
			input.Min(min)
		}
		if max, ok := tags.Lookup("max"); ok {
			input.Max(max)
		}
		if col, ok := tags.Lookup("options"); ok {
			chunks := strings.Split(col, ",")
			var options []html.KeyValue
			for _, option := range chunks {
				parts := strings.Split(option, ":")
				if len(parts) == 2 {
					options = append(options, html.KeyValue{parts[0], parts[1]})
				}
			}
			input.SetOptions(options)
		}
		input.Value = ref.Field(i).Interface()
		frm = append(frm, *input)
	}
	return frm
}

func (c Controller) save(r *evo.Request) {
	if !r.User.HasPerm("settings.access") {
		r.WriteResponse(false, e.Context(constant.ERROR_UNAUTHORIZED))
		return
	}
	name := r.Params("name")
	if !settings.Has(name) {
		r.WriteResponse(false, e.Context(constant.ERROR_OBJECT_NOT_EXIST))
		return
	}
	item := settings.Get(name).(Settings)

	err := r.BodyParser(item.Ptr)
	if err != nil {
		r.WriteResponse(false, err)
		return
	}
	ref.Invoke(item.Ptr, "OnUpdate", r)
	b, err := json.Marshal(item.Ptr)
	item.Data = string(b)

	db.Model(&item).Where("reference = ?", item.Reference).Update("data", item.Data)
	settings.Set(name, item)
	r.Flash("success", i18.T("Successfully saved"))
	r.WriteResponse(true, item)
}

func (c Controller) reset(r *evo.Request) {
	if !r.User.HasPerm("settings.access") {
		r.WriteResponse(false, e.Context(constant.ERROR_UNAUTHORIZED))
		return
	}
	name := r.Params("name")
	if !settings.Has(name) {
		r.WriteResponse(false, e.Context(constant.ERROR_OBJECT_NOT_EXIST))
		return
	}
	item := settings.Get(name).(Settings)

	item.Data = item.Default

	err := json.Unmarshal([]byte(item.Default), item.Ptr)
	if err != nil {
		r.WriteResponse(false, err)
		return
	}
	ref.Invoke(item.Ptr, "OnUpdate", r)
	db.Model(&item).Where("reference = ?", item.Reference).Update("data", item.Data)
	settings.Set(name, item)
	r.WriteResponse(true, item)
}
