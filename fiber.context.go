package evo

import (
	"encoding/json"
	"fmt"
	"github.com/ajg/form"
	"github.com/avct/uasurfer"
	"github.com/getevo/evo/lib/log"
	"github.com/getevo/evo/lib/text"
	"github.com/gofiber/utils"
	"github.com/klauspost/compress/snappy"
	"github.com/valyala/fasthttp"
	"io"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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

// GetBrowser return browser information by parsing useragent
func (r *Request) GetBrowser() *uasurfer.UserAgent {
	return uasurfer.Parse(r.UserAgent())
}

// Accepts checks if the specified extensions or content types are acceptable.
func (r *Request) Accepts(offers ...string) (offer string) {
	return r.Context.Accepts(offers...)
}

// AcceptsCharsets checks if the specified charset is acceptable.
func (r *Request) AcceptsCharsets(offers ...string) (offer string) {
	return r.Context.AcceptsCharsets(offers...)
}

// AcceptsEncodings checks if the specified encoding is acceptable.
func (r *Request) AcceptsEncodings(offers ...string) (offer string) {
	return r.Context.AcceptsEncodings(offers...)
}

// AcceptsLanguages checks if the specified language is acceptable.
func (r *Request) AcceptsLanguages(offers ...string) (offer string) {
	return r.Context.AcceptsLanguages(offers...)
}

// Append the specified value to the HTTP response header field.
// If the header is not already set, it creates the header with the specified value.
func (r *Request) Append(field string, values ...string) {
	r.Context.Append(field, values...)
}

// Attachment sets the HTTP response Content-Disposition header field to attachment.
func (r *Request) Attachment(name ...string) {
	r.Context.Attachment(name...)
}

// BaseURL returns (protocol + host).
func (r *Request) BaseURL() string {
	return r.Context.BaseURL()
}

// Body contains the raw body submitted in a POST request.
func (r *Request) Body() string {
	return string(r.Context.Body())
}

// BodyParser binds the request body to a struct.
// It supports decoding the following content types based on the Content-Type header:
// application/json, application/xml, application/x-www-form-urlencoded, multipart/form-data
func (r *Request) BodyParser(out interface{}) error {
	ctype := string(r.ContentType())

	if strings.HasPrefix(ctype, MIMEApplicationJSON) {
		return json.Unmarshal(r.Context.Context().Request.Body(), out)
	} else if strings.HasPrefix(ctype, MIMETextXML) || strings.HasPrefix(ctype, MIMEApplicationXML) {
		return json.Unmarshal(r.Context.Context().Request.Body(), out)
	} else if strings.HasPrefix(ctype, MIMEApplicationForm) || strings.HasPrefix(ctype, MIMEMultipartForm) {
		return form.DecodeString(out, r.Body())
	}
	return fmt.Errorf("undefined body type")
	//return r.Context.BodyParser(out)
}

// ContentType returns request content type
func (r *Request) ContentType() string {
	return string(r.Context.Context().Request.Header.ContentType())
}

// UserAgent returns request useragent
func (r *Request) UserAgent() string {
	ua := r.Get("User-Agent")
	if ua == "" {
		ua = r.Get("X-Original-Agent")
	}
	if ua == "" {
		ua = r.Get("X-User-Agent")
	}
	return ua
}

// UserAgent returns request useragent
func (r *Request) IP() string {
	clientIP := requestIP(r.Context.Context())
	return clientIP
}

// ClearCookie expires a specific cookie by key.
// If no key is provided it expires all cookies.
func (r *Request) ClearCookie(key ...string) {
	r.Context.ClearCookie(key...)
}

// Cookie sets a cookie by passing a cookie struct
func (r *Request) Cookie(cookie *Cookie) {
	fcookie := fasthttp.AcquireCookie()
	fcookie.SetKey(cookie.Name)
	fcookie.SetValue(cookie.Value)
	fcookie.SetPath(cookie.Path)
	fcookie.SetDomain(cookie.Domain)
	fcookie.SetExpire(cookie.Expires)
	fcookie.SetSecure(cookie.Secure)
	fcookie.SetHTTPOnly(cookie.HTTPOnly)

	switch utils.ToLower(cookie.SameSite) {
	case "strict":
		fcookie.SetSameSite(fasthttp.CookieSameSiteStrictMode)
	case "none":
		fcookie.SetSameSite(fasthttp.CookieSameSiteNoneMode)
	default:
		fcookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	}

	r.Context.Context().Response.Header.SetCookie(fcookie)
	fasthttp.ReleaseCookie(fcookie)
}

// Cookies is used for getting a cookie value by key
func (r *Request) Cookies(key string) (value string) {
	return r.Context.Cookies(key)
}

// Download transfers the file from path as an attachment.
// Typically, browsers will prompt the user for download.
// By default, the Content-Disposition header filename= parameter is the filepath (this typically appears in the browser dialog).
// Override this default with the filename parameter.
func (r *Request) Download(file string, name ...string) {
	r.Context.Download(file, name...)
}

// Format performs content-negotiation on the Accept HTTP header.
// It uses Accepts to select a proper format.
// If the header is not specified or there is no proper format, text/plain is used.
func (r *Request) Format(body interface{}) {
	var b string
	accept := r.Context.Accepts("html", "json")

	switch val := body.(type) {
	case string:
		b = val
	case []byte:
		b = string(val)
	default:
		b = fmt.Sprintf("%+v", val)
	}
	switch accept {
	case "html":
		r.Context.SendString(b)
	case "json":
		if err := r.Context.JSON(body); err != nil {
			log.Error("Format: error serializing json ", err)
		}
	default:
		r.Context.SendString(b)
	}
}

// FormFile returns the first file by key from a MultipartForm.
func (r *Request) FormFile(key string) (*multipart.FileHeader, error) {
	return r.Context.FormFile(key)
}

// FormValue returns the first value by key from a MultipartForm.
func (r *Request) FormValue(key string) (value string) {
	return r.Context.FormValue(key)
}

// FormValue returns the first value by key from a MultipartForm.
func (r *Request) FormValueI(key string, params ...string) value {
	return Value(r.Context.FormValue(key), params...)
}

//Fresh not implemented yet
func (r *Request) Fresh() bool {
	return r.Context.Fresh()
}

// Get returns the HTTP request header specified by field.
// Field names are case-insensitive
func (r *Request) Get(key string) string {
	return r.Context.Get(key)
}

// Get returns the HTTP request header specified by field.
// Field names are case-insensitive
func (r *Request) GetI(key string, params ...string) value {
	return Value(r.Context.Get(key), params...)
}

// Hostname contains the hostname derived from the Host HTTP header.
func (r *Request) Hostname() string {
	if r.Get("X-Forwarded-Host") != "" {
		return r.Get("X-Forwarded-Host")
	}
	if r.Get("X-Forwarded-Server") != "" {
		return r.Get("X-Forwarded-Server")
	}
	return r.Context.Hostname()
}

// IPs returns an string slice of IP addresses specified in the X-Forwarded-For request header.
func (r *Request) IPs() []string {
	if len(r.Context.IPs()) > 0 {
		return r.Context.IPs()
	}
	return []string{}
}

func (r *Request) Header(key string) string {
	return r.Get(key)
}

func (r *Request) Headers(key string) map[string]string {
	var headers map[string]string
	for _, line := range strings.Split(r.Context.Context().Request.Header.String(), "\n") {
		var parts = strings.SplitN(line, ":", 1)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		} else if len(parts) > 0 {
			headers[parts[0]] = ""
		}
	}
	return headers
}

func (r *Request) SetHeader(key, val string) {
	r.Set(key, val)
}

// Is returns the matching content type,
// if the incoming request’s Content-Type HTTP header field matches the MIME type specified by the type parameter
func (r *Request) Is(extension string) (match bool) {
	return r.Context.Is(extension)
}

// JSON converts any interface or string to JSON using Jsoniter.
// This method also sets the content header to application/json.
func (r *Request) JSON(data interface{}) error {
	raw, err := json.Marshal(data)
	// Check for errors
	if err != nil {
		return err
	}
	// Set http headers
	r.Context.Context().Response.Header.SetContentType(MIMEApplicationJSON)
	r.Write(raw)

	return nil
}

// JSONP sends a JSON response with JSONP support.
// This method is identical to JSON, except that it opts-in to JSONP callback support.
// By default, the callback name is simply callback.
func (r *Request) JSONP(json interface{}, callback ...string) error {
	return r.Context.JSONP(json, callback...)
}

// Links joins the links followed by the property to populate the response’s Link HTTP header field.
func (r *Request) Links(link ...string) {
	r.Context.Links(link...)
}

// Locals makes it possible to pass interface{} values under string keys scoped to the request
// and therefore available to all following routes that match the request.
func (r *Request) Locals(key string, value ...interface{}) (val interface{}) {
	return r.Context.Locals(key, value...)
}

// Location sets the response Location HTTP header to the specified path parameter.
func (r *Request) Location(path string) {
	r.Context.Location(path)
}

// Method contains a string corresponding to the HTTP method of the request: GET, POST, PUT and so on.
func (r *Request) Method(override ...string) string {
	return r.Context.Method(override...)
}

// MultipartForm parse form entries from binary.
// This returns a map[string][]string, so given a key the value will be a string slice.
func (r *Request) MultipartForm() (*multipart.Form, error) {
	return r.Context.MultipartForm()
}

// Next executes the next method in the stack that matches the current route.
// You can pass an optional error for custom error handling.
func (r *Request) Next() error {
	return r.Context.Next()
}

// OriginalURL contains the original request URL.
func (r *Request) OriginalURL() string {
	return r.Context.OriginalURL()
}

// Params is used to get the route parameters.
// Defaults to empty string "", if the param doesn't exist.
func (r *Request) Params(key string) string {
	return r.Context.Params(key)
}

// Params is used to get the route parameters.
// Defaults to empty string "", if the param doesn't exist.
func (r *Request) ParamsI(key string, params ...string) value {
	return Value(r.Context.Params(key), params...)
}

// Path returns the path part of the request URL.
// Optionally, you could override the path.
func (r *Request) Path(override ...string) string {
	return r.Context.Path(override...)
}

// Protocol contains the request protocol string: http or https for TLS requests.
func (r *Request) Protocol() string {
	return r.Context.Protocol()
}

// Query returns the query string parameter in the url.
func (r *Request) QueryI(key string, params ...string) value {
	return Value(r.Context.Query(key), params...)
}

// Query returns the query string parameter in the url.
func (r *Request) Query(key string) (value string) {
	return r.Context.Query(key)
}

// Redirect to the URL derived from the specified path, with specified status.
// If status is not specified, status defaults to 302 Found
func (r *Request) Redirect(path string, status ...int) {
	r.Context.Redirect(path, status...)
}

// Render a template with data and sends a text/html response.
// We support the following engines: html, amber, handlebars, mustache, pug
func (r *Request) Render(file string, bind interface{}) error {
	return r.Context.Render(file, bind)
}

// SaveFile saves any multipart file to disk.
func (r *Request) SaveFile(fileheader *multipart.FileHeader, path string) error {
	return r.Context.SaveFile(fileheader, path)
}

// Secure returns a boolean property, that is true, if a TLS connection is established.
func (r *Request) Secure() bool {
	return r.Context.Secure()
}

// Send sets the HTML response body. The Send body can be of any type.
func (r *Request) SendHTML(body interface{}) {
	r.Set("Content-Type", "text/html")
	r.Write(body)
}

// Send sets the HTTP response body. The Send body can be of any type.
func (r *Request) Send(body string) {
	r.Context.Send([]byte(body))
}

// SendBytes sets the HTTP response body for []byte types
// This means no type assertion, recommended for faster performance
func (r *Request) SendBytes(body []byte) {
	r.Context.Send(body)
}

// SendFile transfers the file from the given path.
// The file is compressed by default
// Sets the Content-Type response HTTP header field based on the filenames extension.
func (r *Request) SendFile(file string, noCompression ...bool) {
	r.Context.SendFile(file, noCompression...)
}

// SendStatus sets the HTTP status code and if the response body is empty,
// it sets the correct status message in the body.
func (r *Request) SendStatus(status int) {
	r.Context.SendStatus(status)
}

// SendString sets the HTTP response body for string types
// This means no type assertion, recommended for faster performance
func (r *Request) SendString(body string) {
	r.Context.SendString(body)
}

// Set sets the response’s HTTP header field to the specified key, value.
func (r *Request) Set(key string, val string) {
	r.Context.Set(key, val)
}

// Subdomains returns a string slive of subdomains in the domain name of the request.
// The subdomain offset, which defaults to 2, is used for determining the beginning of the subdomain segments.
func (r *Request) Subdomains(offset ...int) []string {
	return r.Context.Subdomains(offset...)
}

// Stale is not implemented yet, pull requests are welcome!
func (r *Request) Stale() bool {
	return r.Context.Stale()
}

// Status sets the HTTP status for the response.
// This method is chainable.
func (r *Request) Status(status int) *Request {
	r.Context.Status(status)
	return r
}

// Type sets the Content-Type HTTP header to the MIME type specified by the file extension.
func (r *Request) Type(ext string) *Request {
	r.Context.Type(ext)
	return r
}

// Vary adds the given header field to the Vary response header.
// This will append the header, if not already listed, otherwise leaves it listed in the current location.
func (r *Request) Vary(fields ...string) {
	r.Context.Vary(fields...)
}

// Write appends any input to the HTTP body response.
func (r *Request) Write(body interface{}) {
	var cache = true
	var data []byte
	switch body := body.(type) {
	case string:
		data = []byte(body)
	case []byte:
		data = body
	case int:
		data = []byte(strconv.Itoa(body))
	case bool:
		data = []byte(strconv.FormatBool(body))
	case io.Reader:
		r.Context.Context().Response.SetBodyStream(body, -1)
		r.Context.Set(HeaderContentLength, strconv.Itoa(len(r.Context.Context().Response.Body())))
		cache = false
	default:
		data = []byte(fmt.Sprintf("%v", body))
	}
	if cache && r.CacheKey != "" {
		Cache.Set(r.CacheKey, cached{
			content: data,
			header:  r.Context.Context().Response.Header,
			code:    r.Context.Context().Response.StatusCode(),
		}, r.CacheDuration)
	}
	if r.BeforeWrite != nil {
		data = r.BeforeWrite(r, data)
	}
	if r.Header("x-snappy") == "1" {
		data = snappy.Encode(nil, data)
	}
	r.Context.Context().Response.SetBody(data)
}

// XHR returns a Boolean property, that is true, if the request’s X-Requested-With header field is XMLHttpRequest,
// indicating that the request was issued by a client library (such as jQuery).
func (r *Request) XHR() bool {
	return r.Context.XHR()
}

// SetCookie set cookie with given name,value and optional params (wise function)
func (r *Request) SetCookie(key string, val interface{}, params ...interface{}) {
	cookie := new(Cookie)
	cookie.Name = key
	cookie.Path = "/"
	ref := reflect.ValueOf(val)
	switch ref.Kind() {
	case reflect.String:
		cookie.Value = val.(string)
		break
	case reflect.Ptr:
		r.SetCookie(key, ref.Elem().Interface(), params...)
		return
	case reflect.Map, reflect.Struct, reflect.Array, reflect.Slice:
		cookie.Value = text.ToJSON(val)
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
	r.Cookie(cookie)
}

func (r *Request) Unauthorized() bool {
	if r.User.Anonymous {
		log.Error("unauthorized request")
		r.Status(401)
		r.WriteResponse(fmt.Errorf("unauthorized"), 401)
		return true
	}
	return false
}

type Error struct {
	Code    int
	Message string
}

// NewError creates a new HTTPError instance.
func NewError(code int, message ...string) *Error {
	e := &Error{code, utils.StatusMessage(code)}
	if len(message) > 0 {
		e.Message = message[0]
	}
	return e
}
