package evo

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/outcome"
	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/gjson"
)

var errorType = reflect.TypeOf(fmt.Errorf(""))
var errorsType = reflect.TypeOf([]error{})

type CacheControl struct {
	Duration      time.Duration
	Key           string
	ExposeHeaders bool
}

type Request struct {
	Context        *fiber.Ctx
	Response       Response
	CacheControl   *CacheControl
	url            *URL
	status         int
	_break         bool
	jsonParsedBody *gjson.Result
	user           *UserInterface
}

type Response struct {
	Success bool     `json:"success"`
	Error   []string `json:"errors,omitempty"`
	Data    any      `json:"data,omitempty"`
}

type URL struct {
	Query       url.Values
	QueryString string
	Host        string
	Scheme      string
	Path        string
	Raw         string
}

func (response Response) HasError() bool {
	return len(response.Error) > 0
}

func (r *Request) URL() *URL {
	if r.url != nil {
		return r.url
	}
	r.url = &URL{}
	base := strings.Split(r.BaseURL(), "://")
	r.url.Scheme = base[0]
	r.url.Host = r.Hostname()
	r.url.Raw = r.OriginalURL()
	parts := strings.Split(r.url.Raw, "?")
	if len(parts) == 1 {
		r.url.Query = url.Values{}
		r.url.Path = r.url.Raw
		r.url.QueryString = ""
	} else {
		r.url.Path = parts[0]
		r.url.Query, _ = url.ParseQuery(strings.Join(parts[1:], "?"))
		r.url.QueryString = parts[1]
	}
	return r.url
}
func (u *URL) Set(key string, value any) *URL {
	u.Query.Set(key, fmt.Sprint(value))
	return u
}
func (u *URL) String() string {
	return u.Path + "?" + u.Query.Encode()
}

func Upgrade(ctx *fiber.Ctx) *Request {
	r := Request{}
	r.Context = ctx
	r.Response = Response{Success: true}
	return &r
}

func (r *Request) WriteResponse(resp ...any) {
	if len(resp) == 0 {
		return
	}
	var message = false
	for _, item := range resp {
		ref := reflect.ValueOf(item)
		switch ref.Kind() {
		case reflect.Slice:
			if ref.Type() == errorsType {
				r.Response.Success = false
				r.Status(StatusBadRequest)
				for _, err := range item.([]error) {
					r.Response.Error = append(r.Response.Error, err.Error())
				}
				r._writeResponse(r.Response)
				return
			} else {
				if v, ok := item.([]byte); ok {
					r.Write(v)
					return
				}

				r.Response.Success = true
				r.Response.Data = item
				r._writeResponse(r.Response)
				return
			}
		case reflect.Struct, reflect.Ptr:
			if ref.Type() == errorType {
				r.Response.Success = false
				r.Status(StatusBadRequest)
				r.Response.Error = append(r.Response.Error, item.(error).Error())
				r._writeResponse(r.Response)
				return
			}
			for ref.Kind() == reflect.Ptr {
				ref = ref.Elem()
			}
			instance := ref.Interface()

			if v, ok := instance.(Response); ok {
				r.Response = v
				r._writeResponse(r.Response)
				return
			}

			if v, ok := instance.(outcome.Response); ok {

				if v.StatusCode > 0 {
					r.Status(v.StatusCode)
				}
				if len(v.Cookies) > 0 {
					for idx := range v.Cookies {
						r.SetRawCookie(v.Cookies[idx])
					}
				}

				if v.RedirectURL != "" {
					r.Location(v.RedirectURL)
					r._writeResponse(r.Response)
					return
				}

				if v.Data != nil {
					r.Response.Success = true
					r.Response.Data = v.Data
				}

				if v.ContentType != "" {
					r.SetHeader("Content-Type", v.ContentType)
				} else {
					r.SetHeader("Content-Type", fiber.MIMEApplicationJSONCharsetUTF8)
				}

				if len(v.Headers) > 0 {
					for header, value := range v.Headers {
						r.SetHeader(header, value)
					}
				}

				if len(v.Errors) > 0 {
					r.Response.Success = false
					r.Status(StatusBadRequest)
					r.Response.Error = append(r.Response.Error, v.Errors...)
				}
			} else {
				r.Response.Data = item
			}

		case reflect.Bool:
			r.Response.Success = item.(bool)
		case reflect.Int32, reflect.Int16, reflect.Int64:
			r.Response.Data = item.(int)
		case reflect.String:
			if !message {
				r.Response.Data = item.(string)
				message = true
			} else {
				r.Response.Data = item
			}
		default:
			r.Response.Data = item
			r.Response.Success = true
		}

	}
	r._writeResponse(r.Response)
}

func (r *Request) _writeResponse(resp Response) {
	switch v := resp.Data.(type) {
	case []byte:
		r.Write(v)
	default:
		r.JSON(r.Response)
	}

}

func (r *Request) Error(err any, code ...int) bool {
	if err == nil {
		return false
	}
	r._break = true
	if len(code) > 0 {
		r.status = code[0]
	} else if r.status < 400 {
		r.status = fiber.StatusBadRequest
	}
	if v, ok := err.(error); ok {
		r.Response.Error = append(r.Response.Error, v.Error())
	} else {
		r.Response.Error = append(r.Response.Error, fmt.Sprint(err))
	}
	r._writeResponse(r.Response)
	return true
}

func (r *Request) PushError(err any, code ...int) bool {
	if err == nil {
		return false
	}
	if len(code) > 0 {
		r.status = code[0]
	} else if r.status < 400 {
		r.status = fiber.StatusBadRequest
	}
	if v, ok := err.(error); ok {
		r.Response.Error = append(r.Response.Error, v.Error())
		return true
	}
	r.Response.Error = append(r.Response.Error, fmt.Sprint(err))

	return true
}

func (r *Request) HasError() bool {
	return len(r.Response.Error) > 0
}

func (r *Request) Var(key string, value ...any) generic.Value {
	return generic.Parse(r.Context.Locals(key, value...))
}

func (r *Request) RestartRouting() error {
	return r.Context.RestartRouting()
}
