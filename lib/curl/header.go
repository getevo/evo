package curl

import "github.com/getevo/json"

type BasicAuth struct {
	Username string
	Password string
}

type BearerAuth struct {
	Token string
}

// Header represents http request header
type Header map[string]string

func (h Header) Clone() Header {
	if h == nil {
		return nil
	}
	hh := Header{}
	for k, v := range h {
		hh[k] = v
	}
	return hh
}

// ParseStruct parse struct into header
func ParseStruct(h Header, v any) Header {
	data, err := json.Marshal(v)
	if err != nil {
		return h
	}

	err = json.Unmarshal(data, &h)
	return h
}

// HeaderFromStruct init header from struct
func HeaderFromStruct(v any) Header {

	var header Header
	header = ParseStruct(header, v)
	return header
}

type ReservedHeader map[string]string
