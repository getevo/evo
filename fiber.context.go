package evo

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getevo/evo/v2/lib/frm"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/outcome"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/utils/v2"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
	"io"
	"mime"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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

// AppendHeader the specified value to the HTTP response header field.
// If the header is not already set, it creates the header with the specified value.
func (r *Request) AppendHeader(field string, values ...string) {
	r.Context.Append(field, values...)
}

// Attachment sets the HTTP response Content-Disposition header field to attachment.
func (r *Request) Attachment(name ...string) {
	r.Context.Attachment(name...)
}

// QueryString returns url query string.
func (r *Request) QueryString() string {
	return r.URL().QueryString
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
	ctype := r.ContentType()

	if strings.HasPrefix(ctype, MIMEApplicationJSON) {
		return json.Unmarshal(r.Context.Context().Request.Body(), out)
	} else if strings.HasPrefix(ctype, MIMETextXML) || strings.HasPrefix(ctype, MIMEApplicationXML) {
		return json.Unmarshal(r.Context.Context().Request.Body(), out)
	} else if strings.HasPrefix(ctype, MIMEApplicationForm) || strings.HasPrefix(ctype, MIMEMultipartForm) {
		dec := frm.NewDecoder(&frm.DecoderOptions{
			TagName:           "json",
			IgnoreUnknownKeys: true,
		})
		var data, err = r.Form()
		if err != nil {
			return err
		}
		return dec.Decode(data, out)
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
	ua := r.Header("User-Agent")
	if ua == "" {
		ua = r.Header("X-Original-Agent")
	}
	if ua == "" {
		ua = r.Header("X-User-Agent")
	}
	return ua
}

// IP returns user ip
func (r *Request) IP() string {
	if fiberConfig.ServerHeader != "" {
		return r.Header(fiberConfig.ServerHeader)
	}
	return r.Context.IP()
}

// ClearCookie expires a specific cookie by key.
// If no key is provided it expires all cookies.
func (r *Request) ClearCookie(key ...string) {
	r.Context.ClearCookie(key...)
}

// SetRawCookie sets a cookie by passing a cookie struct
func (r *Request) SetRawCookie(cookie *outcome.Cookie) {
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

// Cookie is used for getting a cookie value by key
func (r *Request) Cookie(key string) (value string) {
	return r.Context.Cookies(key)
}

// Download transfers the file from path as an attachment.
// Typically, browsers will prompt the user for download.
// By default, the Content-Disposition header filename= parameter is the filepath (this typically appears in the browser dialog).
// Override this default with the filename parameter.
func (r *Request) Download(file string, name ...string) error {
	return r.Context.Download(file, name...)
}

// Format performs content-negotiation on the Accept HTTP header.
// It uses Accepts to select a proper format.
// If the header is not specified or there is no proper format, text/plain is used.
func (r *Request) Format(body interface{}) error {
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
		return r.Context.SendString(b)
	case "json":
		if err := r.Context.JSON(body); err != nil {
			return err
		}
	default:
		return r.Context.SendString(b)
	}
	return nil
}

// FormFile returns the first file by key from a MultipartForm.
func (r *Request) FormFile(key string) (*multipart.FileHeader, error) {
	return r.Context.FormFile(key)
}

// ParseJsonBody returns parsed JSON Body using gjson.
func (r *Request) ParseJsonBody() *gjson.Result {
	if r.jsonParsedBody == nil {
		var t = gjson.Parse(string(r.Context.Body()))
		r.jsonParsedBody = &t
	}
	return r.jsonParsedBody
}

// BodyValue returns the first value by key from a MultipartForm or JSON Body.
func (r *Request) BodyValue(key string) generic.Value {
	ctype := r.ContentType()
	if strings.HasPrefix(ctype, MIMEApplicationJSON) {
		if r.jsonParsedBody == nil {
			var t = gjson.Parse(string(r.Context.Body()))
			r.jsonParsedBody = &t
		}
		generic.Parse(r.jsonParsedBody.Get(key).Raw)
	} else if strings.HasPrefix(ctype, MIMEApplicationForm) || strings.HasPrefix(ctype, MIMEMultipartForm) {
		return r.FormValue(key)
	}

	return generic.Parse(nil)
}

// FormValue returns the first value by key from a MultipartForm.
func (r *Request) FormValue(key string) generic.Value {
	return generic.Parse(r.Context.FormValue(key))
}

// Form.
func (r *Request) Form() (url.Values, error) {
	var err error
	var form url.Values
	if r.Method() == "POST" || r.Method() == "PUT" || r.Method() == "PATCH" {
		form, err = parsePostForm(r)
	}
	if form == nil {
		form = url.Values{}
	}

	return form, err
}

func parsePostForm(r *Request) (vs url.Values, err error) {
	if r.Body == nil {
		err = errors.New("missing form body")
		return
	}
	ct := r.ContentType()
	// RFC 7231, section 3.1.1.5 - empty type
	//   MAY be treated as application/octet-stream
	if ct == "" {
		ct = "application/octet-stream"
	}
	ct, _, err = mime.ParseMediaType(ct)
	switch {
	case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		return url.ParseQuery(string(r.Body()))
	case strings.HasPrefix(ct, "multipart/form-data"):
		v, err := r.MultipartForm()
		return url.Values(v.Value), err
	}
	return nil, fmt.Errorf("invalid content type")
}

// Fresh not implemented yet
func (r *Request) Fresh() bool {
	return r.Context.Fresh()
}

// Get returns the HTTP request header specified by field.
// Field names are case-insensitive
func (r *Request) Get(key string) generic.Value {
	// It doesn't return POST'ed arguments - use PostArgs() for this.
	//
	// See also PostArgs, FormValue and FormFile.
	if r.Context.Context().QueryArgs().Has(key) {
		return r.Query(key)
	}
	if len(r.Context.Context().Request.Header.Peek(key)) > 0 {
		return generic.Parse(r.Header(key))
	}
	var val = r.Header(key)
	if len(val) > 0 {
		return generic.Parse(val)
	}

	ctype := utils.ToLower(string(r.Context.Context().Request.Header.ContentType()))
	ctype = utils.ParseVendorSpecificContentType(ctype)

	if strings.HasPrefix(ctype, MIMEApplicationForm) {
		val = r.Context.FormValue(key)
		if len(val) > 0 {
			return generic.Parse(val)
		}
	} else if strings.HasPrefix(ctype, MIMEApplicationJSON) {
		return generic.Parse(gjson.Parse(string(r.Context.Body())).Get(key).String())
	}

	return generic.Parse(nil)
}

// Hostname contains the hostname derived from the Host HTTP header.
func (r *Request) Hostname() string {
	if r.Header("X-Forwarded-Host") != "" {
		return r.Header("X-Forwarded-Host")
	}
	if r.Header("X-Forwarded-Server") != "" {
		return r.Header("X-Forwarded-Server")
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
	return r.Context.Get(key)
}

func (r *Request) RespHeaders() map[string]string {
	return r.Context.GetRespHeaders()
}

func (r *Request) ReqHeaders() map[string]string {
	return r.Context.GetReqHeaders()
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

// Param is used to get the route parameters.
// Defaults to empty string "", if the param doesn't exist.
func (r *Request) Param(key string) generic.Value {
	return generic.Parse(r.Context.Params(key))
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
func (r *Request) Query(key string) (value generic.Value) {
	return generic.Parse(r.Context.Query(key))
}

// Redirect to the URL derived from the specified path, with specified status.
// If status is not specified, status defaults to 302 Found
func (r *Request) Redirect(path string, status ...int) error {
	return r.Context.Redirect(path, status...)
}

// SaveFile saves any multipart file to disk.
func (r *Request) SaveFile(fileheader *multipart.FileHeader, path string) error {
	return r.Context.SaveFile(fileheader, path)
}

// IsSecure returns a boolean property, that is true, if a TLS connection is established.
func (r *Request) IsSecure() bool {
	return r.Context.Secure()
}

// Send sets the HTML response body. The Send body can be of any type.
func (r *Request) SendHTML(body interface{}) {
	r.Set("Content-Type", "text/html")
	r.Write(body)
}

// Send sets the HTTP response body. The Send body can be of any type.
func (r *Request) Send(body string) error {
	return r.Context.Send([]byte(body))
}

// SendBytes sets the HTTP response body for []byte types
// This means no type assertion, recommended for faster performance
func (r *Request) SendBytes(body []byte) error {
	return r.Context.Send(body)
}

// SendFile transfers the file from the given path.
// The file is compressed by default
// Sets the Content-Type response HTTP header field based on the filenames extension.
func (r *Request) SendFile(file string, noCompression ...bool) error {
	return r.Context.SendFile(file, noCompression...)
}

// SendStatus sets the HTTP status code and if the response body is empty,
// it sets the correct status message in the body.
func (r *Request) SendStatus(status int) error {
	return r.Context.SendStatus(status)
}

// SendString sets the HTTP response body for string types
// This means no type assertion, recommended for faster performance
func (r *Request) SendString(body string) error {
	return r.Context.SendString(body)
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
	default:
		data = []byte(fmt.Sprintf("%v", body))
	}
	if r.status > 0 {
		r.Status(r.status)
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
	cookie := new(outcome.Cookie)
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
	r.SetRawCookie(cookie)
}

// Params return map of parameters in url
func (r *Request) Params() map[string]string {
	return r.Context.AllParams()
}

// Route generate route for named routes
func (r *Request) Route(name string, params ...interface{}) string {
	var m = fiber.Map{}
	var jump = false
	var route = app.GetRoute(name)
	if len(route.Params) == len(params) {
		for idx, key := range route.Params {
			m[key] = params[idx]
		}
	} else if 2*len(route.Params) == len(params) {
		for idx, param := range params {
			if jump {
				jump = false
				continue
			}
			switch p := param.(type) {
			case string:
				if len(params) > idx+1 {
					m[p] = params[idx+1]
					jump = true
				}
			case map[string]interface{}:
				m = p
			}
		}
	}

	var url, _ = r.Context.GetRouteURL(name, m)
	return url
}

func (r *Request) Break() {
	r._break = true
}
