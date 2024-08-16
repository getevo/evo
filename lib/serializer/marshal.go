package serializer

import (
	"github.com/getevo/json"
	"github.com/kelindar/binary"
)

type Interface struct {
	Marshal   func(v any) ([]byte, error)
	Unmarshal func(data []byte, v any) error
}

func New(marshal func(v any) ([]byte, error), unmarshal func(data []byte, v any) error) Interface {
	return Interface{
		marshal, unmarshal,
	}
}

var JSON = Interface{
	json.Marshal, json.Unmarshal,
}

var Binary = Interface{
	binary.Marshal, binary.Unmarshal,
}
