package curl

import (
	"encoding/xml"
	"fmt"
	"github.com/getevo/json"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

// Resp represents a request with it's response
type Resp struct {
	r      *Req
	req    *http.Request
	resp   *http.Response
	client *http.Client
	cost   time.Duration
	*multipartHelper
	reqBody          []byte
	respBody         []byte
	downloadProgress DownloadProgress
	err              error // delayed error
	json             *gjson.Result
}

type SerializedResponse struct {
	ReqBody  []byte
	RespBody []byte
	Cost     time.Duration
}

func (r *Resp) serialize() SerializedResponse {
	return SerializedResponse{
		ReqBody:  r.reqBody,
		RespBody: r.respBody,
		Cost:     r.cost,
	}
}

func (r *Resp) deserialize(s SerializedResponse) {
	r.reqBody = s.ReqBody
	r.respBody = s.RespBody
	r.cost = s.Cost
}

// Request returns *http.Request
func (r *Resp) Request() *http.Request {
	return r.req
}

// Response returns *http.Response
func (r *Resp) Response() *http.Response {
	return r.resp
}

// Bytes returns response body as []byte
func (r *Resp) Bytes() []byte {
	data, _ := r.ToBytes()
	return data
}

// ToBytes returns response body as []byte,
// return error if error happened when reading
// the response body
func (r *Resp) ToBytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.respBody != nil {
		return r.respBody, nil
	}
	if r.resp != nil {
		defer r.resp.Body.Close()
		respBody, err := io.ReadAll(r.resp.Body)
		if err != nil {
			r.err = err
			return nil, err
		}
		r.respBody = respBody
		return r.respBody, nil
	}
	return nil, nil
}

// String returns response body as string
func (r *Resp) String() string {
	data, _ := r.ToBytes()
	return string(data)
}

// ToString returns response body as string,
// return error if error happened when reading
// the response body
func (r *Resp) ToString() (string, error) {
	data, err := r.ToBytes()
	return string(data), err
}

// ToJSON convert json response body to struct or map
func (r *Resp) ToJSON(v any) error {
	data, err := r.ToBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML convert xml response body to struct or map
func (r *Resp) ToXML(v any) error {
	data, err := r.ToBytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// ToFile download the response body to file with optional download callback
func (r *Resp) ToFile(name string) error {
	//TODO set name to the suffix of url path if name == ""
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	if r.respBody != nil {
		_, err = file.Write(r.respBody)
		return err
	}

	if r.downloadProgress != nil && r.resp.ContentLength > 0 {
		return r.download(file)
	}

	defer r.resp.Body.Close()
	_, err = io.Copy(file, r.resp.Body)
	return err
}

func (r *Resp) download(file *os.File) error {
	p := make([]byte, 1024)
	b := r.resp.Body
	defer b.Close()
	total := r.resp.ContentLength
	var current int64
	var lastTime time.Time

	defer func() {
		r.downloadProgress(current, total)
	}()

	for {
		l, err := b.Read(p)
		if l > 0 {
			_, _err := file.Write(p[:l])
			if _err != nil {
				return _err
			}
			current += int64(l)
			if now := time.Now(); now.Sub(lastTime) > r.r.progressInterval {
				lastTime = now
				r.downloadProgress(current, total)
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

var regNewline = regexp.MustCompile(`\n|\r`)

func (r *Resp) autoFormat(s fmt.State) {
	req := r.req
	if r.r.flag&Lcost != 0 {
		fmt.Fprint(s, req.Method, " ", req.URL.String(), " ", r.cost)
	} else {
		fmt.Fprint(s, req.Method, " ", req.URL.String())
	}

	// test if it is should be outputed pretty
	var pretty bool
	var parts []string
	addPart := func(part string) {
		if part == "" {
			return
		}
		parts = append(parts, part)
		if !pretty && regNewline.MatchString(part) {
			pretty = true
		}
	}
	if r.r.flag&LreqBody != 0 { // request body
		addPart(string(r.reqBody))
	}
	if r.r.flag&LrespBody != 0 { // response body
		addPart(r.String())
	}

	for _, part := range parts {
		if pretty {
			fmt.Fprint(s, "\n")
		}
		fmt.Fprint(s, " ", part)
	}
}

func (r *Resp) miniFormat(s fmt.State) {
	req := r.req
	if r.r.flag&Lcost != 0 {
		fmt.Fprint(s, req.Method, " ", req.URL.String(), " ", r.cost)
	} else {
		fmt.Fprint(s, req.Method, " ", req.URL.String())
	}
	if r.r.flag&LreqBody != 0 && len(r.reqBody) > 0 { // request body
		str := regNewline.ReplaceAllString(string(r.reqBody), " ")
		fmt.Fprint(s, " ", str)
	}
	if r.r.flag&LrespBody != 0 && r.String() != "" { // response body
		str := regNewline.ReplaceAllString(r.String(), " ")
		fmt.Fprint(s, " ", str)
	}
}
func (r *Resp) Status() int {
	return r.Response().StatusCode
}

func (r *Resp) Dot(input string) gjson.Result {
	if r.json == nil {
		var j = gjson.Parse(r.String())
		r.json = &j
	}
	return r.json.Get(input)
}
