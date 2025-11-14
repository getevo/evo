package outcome

import (
	"encoding/base64"
	"fmt"
	"github.com/getevo/json"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v2"
)

type HTTPSerializer interface {
	GetResponse() Response
}

type Response struct {
	ContentType string
	Data        any
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

func Json(input any) *Response {
	var response = Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  200,
	}
	response.Data, _ = json.Marshal(input)
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

func (response *Response) Content(input any) *Response {
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

func (response *Response) Cookie(key string, val any, params ...any) *Response {
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

func (response *Response) Error(value any, code ...int) *Response {
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

func (response *Response) SetCacheControl(t time.Duration, headers ...string) *Response {
	var ccHeader string = fmt.Sprintf("max-age=%.0f", t.Seconds())
	var options string

	for _, header := range headers {
		options = options + fmt.Sprintf(", %s", header)
	}

	ccHeader = ccHeader + options

	if response.Headers == nil {
		response.Headers = map[string]string{}
	}
	response.Headers["Cache-Control"] = ccHeader

	return response
}

// processResponseData processes the input data and returns the appropriate byte representation
func processResponseData(input any) []byte {
	if input == nil {
		return nil
	}

	switch v := input.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	default:
		// Check if it's a struct
		val := reflect.ValueOf(input)
		kind := val.Kind()
		if kind == reflect.Ptr {
			if val.IsNil() {
				return nil
			}
			kind = val.Elem().Kind()
		}

		if kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array {
			data, err := json.Marshal(input)
			if err != nil {
				return []byte(fmt.Sprint(input))
			}
			return data
		}

		return []byte(fmt.Sprint(input))
	}
}

// BadRequest returns a 400 Bad Request response
func BadRequest(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusBadRequest,
	}

	if len(input) == 0 {
		response.Data = []byte("Bad Request")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// InternalServerError returns a 500 Internal Server Error response
func InternalServerError(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusInternalServerError,
	}

	if len(input) == 0 {
		response.Data = []byte("Internal Server Error")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// UnAuthorized returns a 401 Unauthorized response
func UnAuthorized(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusUnauthorized,
	}

	if len(input) == 0 {
		response.Data = []byte("Unauthorized")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// StatusOk returns a 200 OK response
func StatusOk(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusOK,
	}

	if len(input) == 0 {
		response.Data = []byte("OK")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// NoContent returns a 204 No Content response
func NoContent(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusNoContent,
	}

	if len(input) == 0 {
		response.Data = []byte("")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// NotFound returns a 404 Not Found response
func NotFound(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusNotFound,
	}

	if len(input) == 0 {
		response.Data = []byte("Not Found")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// NotAcceptable returns a 406 Not Acceptable response
func NotAcceptable(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusNotAcceptable,
	}

	if len(input) == 0 {
		response.Data = []byte("Not Acceptable")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// RequestTimeout returns a 408 Request Timeout response
func RequestTimeout(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusRequestTimeout,
	}

	if len(input) == 0 {
		response.Data = []byte("Request Timeout")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// TooManyRequests returns a 429 Too Many Requests response
func TooManyRequests(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusTooManyRequests,
	}

	if len(input) == 0 {
		response.Data = []byte("Too Many Requests")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// UnavailableForLegalReasons returns a 451 Unavailable For Legal Reasons response
func UnavailableForLegalReasons(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusUnavailableForLegalReasons,
	}

	if len(input) == 0 {
		response.Data = []byte("Unavailable For Legal Reasons")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// Created returns a 201 Created response
func Created(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusCreated,
	}

	if len(input) == 0 {
		response.Data = []byte("Created")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}

// Accepted returns a 202 Accepted response
func Accepted(input ...any) *Response {
	response := &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  fiber.StatusAccepted,
	}

	if len(input) == 0 {
		response.Data = []byte("Accepted")
	} else {
		response.Data = processResponseData(input[0])
	}

	return response
}
