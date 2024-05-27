package errors

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/outcome"
	"github.com/getevo/evo/v2/lib/text"
)

type HTTPError outcome.Response

var (
	NotFound                    = New("Not Found", 404)
	BadRequest                  = New("Bad Request", 400)
	Unauthorized                = New("Unauthorized", 401)
	Forbidden                   = New("Forbidden", 403)
	Internal                    = New("Internal Server Error", 500)
	NotAcceptable               = New("Not Acceptable", 406)
	Conflict                    = New("Conflict", 409)
	Precondition                = New("Precondition Failed", 412)
	UnsupportedMedia            = New("Unsupported Media Type", 415)
	Gone                        = New("Gone", 410)
	RequestTimeout              = New("Request Timeout", 408)
	RequestEntityTooLarge       = New("Request Entity Too Large", 413)
	RequestURITooLong           = New("Request URI Too Long", 414)
	RequestHeaderFieldsTooLarge = New("Request Header Fields Too Large", 431)
	UnavailableForLegalReasons  = New("Unavailable For Legal Reasons", 451)
	PayloadTooLarge             = New("Payload Too Large", 413)
	TooManyRequests             = New("Too Many Requests", 429)
)

func New(err ...interface{}) HTTPError {
	var r = Response{
		Success: false,
		Error:   "Internal Server Error",
	}
	var out = HTTPError{
		StatusCode: 500,
	}

	for i, _ := range err {
		switch v := err[i].(type) {
		case string:
			r.Error = v
		case error:
			r.Error = v.Error()
		case int:
			out.StatusCode = v
		default:
			r.Error = fmt.Sprint(v)
		}
	}
	out.Data = text.ToJSON(r)
	return out
}

func (e HTTPError) Code(code int) HTTPError {
	e.StatusCode = code
	return e
}

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
