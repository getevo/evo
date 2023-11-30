package kv

import "time"

type Params struct {
	Duration time.Duration
	Bucket   string
}
type bucket string

func Bucket(s string) bucket {
	return bucket(s)
}

func Parse(params []any) Params {
	var p = Params{
		Duration: time.Duration(-1),
	}
	for idx, _ := range params {
		switch v := params[idx].(type) {
		case Params:
			return v
		case time.Duration:
			p.Duration = v
		case string:
			p.Bucket = v
		case bucket:
			p.Bucket = string(v)
		}
	}

	return p
}
