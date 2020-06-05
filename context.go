package evo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/CloudyKit/jet"
	"github.com/gofiber/fiber"
	"github.com/iesreza/io/errors"
	"github.com/iesreza/io/lib/jwt"
	"github.com/iesreza/io/lib/log"
	"github.com/iesreza/io/user"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type Request struct {
	Variables fiber.Map
	Context   *fiber.Ctx
	JWT       *jwt.Payload
	User      *user.User
	Response  Response
	flashes   []flash
}
type flash struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   e.Errors    `json:"errors"`
	Data    interface{} `json:"data"`
	Code    int         `json:"code"`
}

type URL struct {
	Query  url.Values
	Host   string
	Scheme string
	Path   string
	Raw    string
}

func (response Response) HasError() bool {
	return response.Error.Exist()
}

func (r *Request) URL() *URL {
	u := URL{}
	base := strings.Split(r.BaseURL(), "://")
	u.Scheme = base[0]
	u.Host = r.Hostname()
	u.Raw = r.OriginalURL()
	parts := strings.Split(u.Raw, "?")
	if len(parts) == 1 {
		u.Query = url.Values{}
		u.Path = u.Raw
	} else {
		u.Path = parts[0]
		u.Query, _ = url.ParseQuery(strings.Join(parts[1:], "?"))
	}
	return &u
}
func (u *URL) Set(key string, value interface{}) *URL {
	u.Query.Set(key, fmt.Sprint(value))
	return u
}
func (u *URL) String() string {
	return u.Path + "?" + u.Query.Encode()
}

func Upgrade(ctx *fiber.Ctx) *Request {
	r := Request{}
	r.Variables = fiber.Map{}
	r.Context = ctx
	r.Response = Response{}
	r.Response.Error = e.Errors{}

	if r.Cookies("access_token") != "" {
		token, err := jwt.Verify(r.Cookies("access_token"))
		if err == nil {
			r.JWT = &token
			r.User = getUser(&token)
		} else {
			r.SetCookie("access_token", "")
			r.Status(http.StatusUnauthorized)
			r.Send("invalid JWT token")
			log.Error(err)
		}
	} else {
		r.JWT = &jwt.Payload{Empty: true, Data: map[string]interface{}{}}
	}
	if r.User == nil {
		r.User = &user.User{Anonymous: true}
	}
	return &r
}

func getUser(payload *jwt.Payload) *user.User {
	var user user.User
	// return user using jwt
	if payload.Data != nil {
		if id, ok := payload.Data["id"]; ok {
			Database.Where("id = ?", id).Take(&user)
			if user.ID == 0 {
				user.Anonymous = true
			}
		}
	}

	return &user
}

func (r *Request) Flash(params ...string) {
	if len(params) == 0 {
		return
	}
	if len(r.flashes) == 0 {
		cookie := r.Cookies("flash")
		if cookie != "" {
			json.Unmarshal([]byte(cookie), &r.flashes)
		}
	}
	if len(params) == 1 {
		r.flashes = append(r.flashes, flash{"info", params[0]})

	} else {
		r.flashes = append(r.flashes, flash{params[0], params[1]})
	}
	r.SetCookie("flash", r.flashes)
}

func (r *Request) Persist() {
	if !r.JWT.Empty {
		exp := time.Now().Add(config.JWT.Age)
		if d, exist := r.JWT.Get("_extend_duration"); exist {
			duration := d.(time.Duration)
			exp = time.Now().Add(duration)
		}
		token, err := jwt.Generate(r.JWT.Data)
		if err == nil {
			r.Cookie(&fiber.Cookie{
				Name:    "access_token",
				Value:   token,
				Expires: exp,
			})

		} else {
			log.Error(err)
		}

	}
}

func (r *Request) View(data interface{}, views ...string) {
	buff := r.RenderView(data, views...)
	buff.Bytes()
	r.SendHTML(buff.Bytes())
}

func (r *Request) RenderView(input interface{}, views ...string) *bytes.Buffer {
	var buff bytes.Buffer
	vars := jet.VarMap{}
	vars.Set("base", r.Context.Protocol()+"://"+r.Context.Hostname())
	vars.Set("proto", r.Context.Protocol())
	vars.Set("hostname", r.Context.Hostname())
	vars.Set("request", r)
	ref := reflect.ValueOf(input)
	kind := ref.Kind()
	if kind == reflect.Map {
		for _, k := range ref.MapKeys() {
			vars.Set(k.String(), ref.MapIndex(k).Interface())
		}
	} else if kind == reflect.String {
		vars.Set("body", input.(string))
	} else {
		vars.Set("param", input)
	}

	for k, v := range r.Variables {
		vars.Set(k, v)
	}

	for _, view := range views {
		buff = bytes.Buffer{}
		parts := strings.Split(view, ".")

		if len(parts) > 1 {
			t, err := GetView(parts[0], strings.Join(parts[1:], "."))
			if err == nil {
				t.Execute(&buff, vars, map[string]interface{}{})
			} else {
				log.Error(err)
				log.Error(parts)
			}
			vars.Set("body", buff.Bytes())

		}
	}
	return &buff
}

func (r *Request) WriteResponse(resp ...interface{}) {
	if len(resp) == 0 {
		r._writeResponse(r.Response)
		return
	}

	var message = false
	for _, item := range resp {
		ref := reflect.ValueOf(item)

		switch ref.Kind() {
		case reflect.Struct:
			if v, ok := item.(Response); ok {
				r._writeResponse(v)
				return
			} else if v, ok := item.(e.Error); ok {
				r.Response.Error.Push(&v)
				r._writeResponse(r.Response)
			} else {
				r.Response.Data = item
			}

		case reflect.Ptr:
			obj := ref.Elem().Interface()
			if v, ok := obj.(Response); ok {
				r._writeResponse(v)
				return
			} else if v, ok := obj.(e.Error); ok {
				r.Response.Error.Push(&v)
			} else if v, ok := item.(error); ok {

				r.Response.Error.Push(e.Context(v.Error()))
			} else {
				r.Response.Data = obj
			}

			break
		case reflect.Bool:
			r.Response.Success = item.(bool)
			break
		case reflect.Int32:
			if r.Response.Code == 0 {
				r.Response.Code = item.(int)
			} else {
				r.Response.Data = item.(int)
			}
			break
		case reflect.String:
			if !message {
				r.Response.Message = item.(string)
				message = true
			} else {
				r.Response.Data = item
			}
			break
		default:
			r.Response.Data = item
		}

	}
	r._writeResponse(r.Response)

}

func (r *Request) _writeResponse(resp Response) {
	if resp.HasError() {
		r.Response.Success = false
	} else {
		r.Response.Success = true
	}
	r.JSON(resp)
}

func (r *Request) SetError(err interface{}) {
	if v, ok := err.(error); ok {
		r.Response.Error.Push(e.Context(v.Error()))
		return
	}
	if v, ok := err.(e.Error); ok {
		r.Response.Error.Push(&v)
		return
	}
	if v, ok := err.(*e.Error); ok {
		r.Response.Error.Push(v)
		return
	}
	log.Error("invalid error provided %+v", err)

}

func (r *Request) Throw(e *e.Error) {
	r.Response.Error.Push(e)
	r.WriteResponse()
}

func (r *Request) HasError() bool {
	return r.Response.Error.Exist()
}

func (r *Request) Var(key string, value interface{}) {
	r.Variables[key] = value
}

func (r Request) GetFlashes() []flash {
	var resp []flash
	flash := r.Cookies("flash")
	if flash != "" {
		json.Unmarshal([]byte(flash), &resp)
	}

	return resp
}
