package outcome

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"reflect"
	"time"
)

type Response struct {
	ContentType string
	Data        interface{}
	StatusCode  int
	Headers     map[string]string
	RedirectURL string
	Cookies     []*Cookie
	Errors      []string
}

func Text(input string) *Response {
	var response = Response{
		ContentType: fiber.MIMETextPlain,
		StatusCode:  200,
		Data:        []byte(input),
	}
	return &response
}

func Html(input string) *Response {
	var response = Response{
		ContentType: fiber.MIMETextHTMLCharsetUTF8,
		StatusCode:  200,
		Data:        []byte(input),
	}
	return &response
}

func Redirect(to string, code ...int) *Response {
	response := &Response{}
	response.RedirectURL = to
	if len(code) > 0 {
		response.StatusCode = code[0]
	} else {
		response.StatusCode = fiber.StatusTemporaryRedirect
	}
	return response
}
func RedirectPermanent(to string) *Response {
	response := &Response{}
	return response.Redirect(to, fiber.StatusPermanentRedirect)
}
func RedirectTemporary(to string) *Response {
	response := &Response{}
	return response.Redirect(to, fiber.StatusTemporaryRedirect)
}

func Json(input interface{}) *Response {
	var response = Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  200,
	}
	response.Data = input
	return &response
}

func (response *Response) Status(code int) *Response {
	response.StatusCode = code
	return response
}

func (response *Response) Header(key, value string) *Response {
	if response.Headers == nil {
		response.Headers = map[string]string{}
	}
	response.Headers[key] = value
	return response
}

func (response *Response) Content(input interface{}) *Response {
	switch v := input.(type) {
	case string:
		response.Data = []byte(v)
	case []byte:
		response.Data = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		response.Data = []byte(fmt.Sprint(v))
	default:
		var err error
		response.Data, err = json.Marshal(input)
		if err != nil {
			response.StatusCode = 400
			response.Data = []byte(err.Error())
		}

	}

	return response
}

// Cookie data for ctx.Cookie
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Path     string    `json:"path"`
	Domain   string    `json:"domain"`
	Expires  time.Time `json:"expires"`
	Secure   bool      `json:"secure"`
	HTTPOnly bool      `json:"http_only"`
	SameSite string    `json:"same_site"`
}

func (response *Response) Cookie(key string, val interface{}, params ...interface{}) *Response {
	cookie := new(Cookie)
	cookie.Name = key
	cookie.Path = "/"
	ref := reflect.ValueOf(val)
	switch ref.Kind() {
	case reflect.String:
		cookie.Value = val.(string)
		break
	case reflect.Ptr:
		response.Cookie(key, ref.Elem().Interface(), params...)
		return response
	case reflect.Map, reflect.Struct, reflect.Array, reflect.Slice:
		b, _ := json.Marshal(val)
		cookie.Value = base64.RawStdEncoding.EncodeToString(b)
		break
	default:
		cookie.Value = fmt.Sprint(val)
	}

	for _, item := range params {
		if v, ok := item.(time.Duration); ok {
			cookie.Expires = time.Now().Add(v)
		}
		if v, ok := item.(time.Time); ok {
			cookie.Expires = v
		}
	}
	response.Cookies = append(response.Cookies, cookie)
	return response
}

func (response *Response) RawCookie(cookie Cookie) *Response {
	response.Cookies = append(response.Cookies, &cookie)
	return response
}

func (response *Response) Redirect(to string, code ...int) *Response {
	response.RedirectURL = to
	if len(code) > 0 {
		response.StatusCode = code[0]
	} else {
		response.StatusCode = fiber.StatusTemporaryRedirect
	}
	return response
}
func (response *Response) RedirectPermanent(to string) *Response {
	return response.Redirect(to, fiber.StatusPermanentRedirect)
}
func (response *Response) RedirectTemporary(to string) *Response {
	return response.Redirect(to, fiber.StatusTemporaryRedirect)
}

func (response *Response) Error(value interface{}, code ...int) *Response {
	if len(code) > 0 {
		response.StatusCode = code[0]
	} else {
		response.StatusCode = fiber.StatusBadRequest
	}
	response.Errors = append(response.Errors, fmt.Sprint(value))
	return response
}

func (response *Response) ShowInBrowser() *Response {
	return response.Header("Content-Disposition", "inline")
}

func (response *Response) Filename(filename string) *Response {
	return response.Header("Content-Disposition", "attachment' filename=\""+filename+"\"")
}

func (response *Response) ResponseSerializer() *Response {
	return response
}
