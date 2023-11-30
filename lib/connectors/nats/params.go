package nats

type Params struct {
	JetStream    bool
	Bucket       string
	IgnorePrefix bool
}
type bucket string

const IgnorePrefix = 1

func Bucket(s string) bucket {
	return bucket(s)
}

func Parse(params []any) Params {
	var p = Params{
		JetStream: false,
	}
	for idx, _ := range params {
		switch v := params[idx].(type) {
		case int:
			if v == IgnorePrefix {
				p.IgnorePrefix = true
			}
		case Params:
			return v
		case string:
			p.Bucket = v
			p.JetStream = true
		case bucket:
			p.JetStream = true
			p.Bucket = string(v)
		}
	}

	return p
}
