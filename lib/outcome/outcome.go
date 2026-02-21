package outcome

import (
	"encoding/base64"
	"fmt"
	"github.com/getevo/json"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v3"
)

// HTTPSerializer is an interface for types that can serialize to HTTP responses
type HTTPSerializer interface {
	GetResponse() Response
}

// Response represents an HTTP response with all its components
type Response struct {
	ContentType string
	Data        interface{} // Accepts []byte or string, automatically converted to []byte when needed
	StatusCode  int
	Headers     map[string]string
	RedirectURL string
	Cookies     []*Cookie
	Errors      []string // Collected errors for structured error responses
}

// Text returns a plain text response with 200 OK status
func Text(input string) *Response {
	return &Response{
		ContentType: fiber.MIMETextPlain,
		StatusCode:  200,
		Data:        []byte(input),
	}
}

// Html returns an HTML response with 200 OK status
func Html(input string) *Response {
	return &Response{
		ContentType: fiber.MIMETextHTMLCharsetUTF8,
		StatusCode:  200,
		Data:        []byte(input),
	}
}

// Redirect returns a redirect response with optional status code (default 307 Temporary Redirect)
func Redirect(to string, code ...int) *Response {
	response := &Response{
		RedirectURL: to,
	}
	if len(code) > 0 {
		response.StatusCode = code[0]
	} else {
		response.StatusCode = fiber.StatusTemporaryRedirect
	}
	return response
}

// RedirectPermanent returns a 301 Permanent Redirect response
func RedirectPermanent(to string) *Response {
	return &Response{
		RedirectURL: to,
		StatusCode:  fiber.StatusPermanentRedirect,
	}
}

// RedirectTemporary returns a 307 Temporary Redirect response
func RedirectTemporary(to string) *Response {
	return &Response{
		RedirectURL: to,
		StatusCode:  fiber.StatusTemporaryRedirect,
	}
}

// Json returns a JSON response with 200 OK status
// If marshaling fails, returns a 500 Internal Server Error
func Json(input any) *Response {
	data, err := json.Marshal(input)
	if err != nil {
		return InternalServerError(fmt.Sprintf("JSON marshaling error: %v", err))
	}
	return &Response{
		ContentType: fiber.MIMEApplicationJSONCharsetUTF8,
		StatusCode:  200,
		Data:        data,
	}
}

// Status sets the HTTP status code for the response
func (response *Response) Status(code int) *Response {
	response.StatusCode = code
	return response
}

// Header adds or updates a response header
func (response *Response) Header(key, value string) *Response {
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers[key] = value
	return response
}

// Content sets the response body content, automatically determining the best serialization
func (response *Response) Content(input any) *Response {
	switch v := input.(type) {
	case string:
		response.Data = []byte(v)
	case []byte:
		response.Data = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		response.Data = []byte(fmt.Sprint(v))
	default:
		data, err := json.Marshal(input)
		if err != nil {
			response.StatusCode = 500
			response.Data = []byte(fmt.Sprintf("Content serialization error: %v", err))
		} else {
			response.Data = data
			// Set JSON content type if not already set
			if response.ContentType == "" {
				response.ContentType = fiber.MIMEApplicationJSONCharsetUTF8
			}
		}
	}
	return response
}

// Cookie represents an HTTP cookie configuration
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

// Cookie adds a cookie to the response. Complex types (maps, structs, slices) are JSON-encoded and base64-encoded.
// Optional params can include time.Duration or time.Time for expiration.
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

// RawCookie adds a pre-configured cookie to the response
func (response *Response) RawCookie(cookie Cookie) *Response {
	response.Cookies = append(response.Cookies, &cookie)
	return response
}

// Redirect sets the redirect URL and optional status code (default 307 Temporary Redirect)
func (response *Response) Redirect(to string, code ...int) *Response {
	response.RedirectURL = to
	if len(code) > 0 {
		response.StatusCode = code[0]
	} else {
		response.StatusCode = fiber.StatusTemporaryRedirect
	}
	return response
}

// RedirectPermanent sets a 301 Permanent Redirect
func (response *Response) RedirectPermanent(to string) *Response {
	return response.Redirect(to, fiber.StatusPermanentRedirect)
}

// RedirectTemporary sets a 307 Temporary Redirect
func (response *Response) RedirectTemporary(to string) *Response {
	return response.Redirect(to, fiber.StatusTemporaryRedirect)
}

// Error adds an error message to the response and sets status code (default 400 Bad Request)
// The error is both added to the Errors slice and set as response data
func (response *Response) Error(value any, code ...int) *Response {
	if len(code) > 0 {
		response.StatusCode = code[0]
	} else {
		response.StatusCode = fiber.StatusBadRequest
	}
	errMsg := fmt.Sprint(value)
	response.Errors = append(response.Errors, errMsg)

	// If no data set yet, use the error message as response body
	if response.Data == nil || len(response.GetData()) == 0 {
		response.Data = []byte(errMsg)
	}
	return response
}

// ShowInBrowser sets Content-Disposition to inline for browser display
func (response *Response) ShowInBrowser() *Response {
	return response.Header("Content-Disposition", "inline")
}

// Filename sets Content-Disposition header for file download with specified filename
func (response *Response) Filename(filename string) *Response {
	return response.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
}

// ResponseSerializer returns the response itself (implements HTTPSerializer interface)
func (response *Response) ResponseSerializer() *Response {
	return response
}

// GetData returns the response data as []byte, automatically converting string to []byte if needed
func (response *Response) GetData() []byte {
	switch v := response.Data.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	case nil:
		return nil
	default:
		// For other types, try to marshal as JSON
		data, err := json.Marshal(v)
		if err != nil {
			return []byte(fmt.Sprint(v))
		}
		return data
	}
}

// SetCacheControl sets the Cache-Control header with max-age and optional directives
func (response *Response) SetCacheControl(t time.Duration, headers ...string) *Response {
	ccHeader := fmt.Sprintf("max-age=%.0f", t.Seconds())

	for _, header := range headers {
		ccHeader += fmt.Sprintf(", %s", header)
	}

	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["Cache-Control"] = ccHeader

	return response
}

// processResponseData processes the input data and returns the appropriate byte representation and content type
func processResponseData(input any) ([]byte, string) {
	if input == nil {
		return nil, fiber.MIMETextPlainCharsetUTF8
	}

	switch v := input.(type) {
	case string:
		return []byte(v), fiber.MIMETextPlainCharsetUTF8
	case []byte:
		return v, "application/octet-stream"
	default:
		// Check if it's a complex type that should be JSON
		val := reflect.ValueOf(input)
		kind := val.Kind()
		if kind == reflect.Ptr {
			if val.IsNil() {
				return nil, fiber.MIMETextPlainCharsetUTF8
			}
			kind = val.Elem().Kind()
		}

		if kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array {
			data, err := json.Marshal(input)
			if err != nil {
				return []byte(fmt.Sprint(input)), fiber.MIMETextPlainCharsetUTF8
			}
			return data, fiber.MIMEApplicationJSONCharsetUTF8
		}

		return []byte(fmt.Sprint(input)), fiber.MIMETextPlainCharsetUTF8
	}
}

// newResponse is a factory function to create status responses with consistent behavior
func newResponse(status int, defaultMsg string, input ...any) *Response {
	response := &Response{
		StatusCode: status,
	}

	if len(input) == 0 {
		response.Data = []byte(defaultMsg)
		response.ContentType = fiber.MIMETextPlainCharsetUTF8
	} else {
		response.Data, response.ContentType = processResponseData(input[0])
	}

	return response
}

// 2xx Success Responses

// OK returns a 200 OK response
func OK(input ...any) *Response {
	return newResponse(fiber.StatusOK, "OK", input...)
}

// Created returns a 201 Created response
func Created(input ...any) *Response {
	return newResponse(fiber.StatusCreated, "Created", input...)
}

// Accepted returns a 202 Accepted response
func Accepted(input ...any) *Response {
	return newResponse(fiber.StatusAccepted, "Accepted", input...)
}

// NoContent returns a 204 No Content response
func NoContent(input ...any) *Response {
	return newResponse(fiber.StatusNoContent, "", input...)
}

// 4xx Client Error Responses

// BadRequest returns a 400 Bad Request response
func BadRequest(input ...any) *Response {
	return newResponse(fiber.StatusBadRequest, "Bad Request", input...)
}

// Unauthorized returns a 401 Unauthorized response
func Unauthorized(input ...any) *Response {
	return newResponse(fiber.StatusUnauthorized, "Unauthorized", input...)
}

// UnAuthorized is deprecated. Use Unauthorized instead.
func UnAuthorized(input ...any) *Response {
	return Unauthorized(input...)
}

// Forbidden returns a 403 Forbidden response
func Forbidden(input ...any) *Response {
	return newResponse(fiber.StatusForbidden, "Forbidden", input...)
}

// NotFound returns a 404 Not Found response
func NotFound(input ...any) *Response {
	return newResponse(fiber.StatusNotFound, "Not Found", input...)
}

// NotAcceptable returns a 406 Not Acceptable response
func NotAcceptable(input ...any) *Response {
	return newResponse(fiber.StatusNotAcceptable, "Not Acceptable", input...)
}

// RequestTimeout returns a 408 Request Timeout response
func RequestTimeout(input ...any) *Response {
	return newResponse(fiber.StatusRequestTimeout, "Request Timeout", input...)
}

// Conflict returns a 409 Conflict response
func Conflict(input ...any) *Response {
	return newResponse(fiber.StatusConflict, "Conflict", input...)
}

// UnprocessableEntity returns a 422 Unprocessable Entity response
func UnprocessableEntity(input ...any) *Response {
	return newResponse(fiber.StatusUnprocessableEntity, "Unprocessable Entity", input...)
}

// TooManyRequests returns a 429 Too Many Requests response
func TooManyRequests(input ...any) *Response {
	return newResponse(fiber.StatusTooManyRequests, "Too Many Requests", input...)
}

// UnavailableForLegalReasons returns a 451 Unavailable For Legal Reasons response
func UnavailableForLegalReasons(input ...any) *Response {
	return newResponse(fiber.StatusUnavailableForLegalReasons, "Unavailable For Legal Reasons", input...)
}

// 5xx Server Error Responses

// InternalServerError returns a 500 Internal Server Error response
func InternalServerError(input ...any) *Response {
	return newResponse(fiber.StatusInternalServerError, "Internal Server Error", input...)
}

// ServiceUnavailable returns a 503 Service Unavailable response
func ServiceUnavailable(input ...any) *Response {
	return newResponse(fiber.StatusServiceUnavailable, "Service Unavailable", input...)
}

// GatewayTimeout returns a 504 Gateway Timeout response
func GatewayTimeout(input ...any) *Response {
	return newResponse(fiber.StatusGatewayTimeout, "Gateway Timeout", input...)
}

// Deprecated: Use OK instead
func StatusOk(input ...any) *Response {
	return OK(input...)
}
