package kafka

type Params struct {
	IgnorePrefix bool
}

const IgnorePrefix = 1

func Parse(params []any) Params {
	var p = Params{}
	for idx, _ := range params {
		switch v := params[idx].(type) {
		case int:
			if v == IgnorePrefix {
				p.IgnorePrefix = true
			}
		case Params:
			return v
		}
	}

	return p
}
