package rdb

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/go-playground/validator"
)

type Source uint8

const (
	Get      Source = 0
	Post     Source = 1
	URL      Source = 3
	Header   Source = 4
	Cookie   Source = 5
	Any      Source = 6
	Constant Source = 7
)

var validate = validator.New()

type Parser struct {
	Params    []Param
	Processor func(params []string) []string
}
type Param struct {
	Key        string
	Source     Source
	Validation string
	Default    string
}

func NewParser(params ...Param) *Parser {
	parser := Parser{}
	parser.Params = params
	return &parser
}

func (parser *Parser) Parse(r *evo.Request) ([]string, error) {
	var data string
	var res []string
	var err error
	var dataMap = map[string]string{}
	var isForm = false
	if r.BodyParser(&dataMap) != nil {
		isForm = true
	}

	for _, item := range parser.Params {
		switch item.Source {
		case Constant:
			res = append(res, item.Default)
			continue
			break
		case Get:
			data = r.Query(item.Key)
			break
		case Post:
			if isForm {
				data = r.FormValue(item.Key)
			} else {
				if v, ok := dataMap[item.Key]; ok {
					data = v
				} else {
					data = ""
				}
			}
			break
		case Header:
			data = r.Get(item.Key)
			break
		case URL:
			data = r.Params(item.Key)
			break
		case Cookie:
			data = r.Cookies(item.Key)
			break
		case Any:
			data = r.Query(item.Key)
			if data == "" {
				if v, ok := dataMap[item.Key]; ok {
					data = v
				} else {
					data = ""
				}
				if data == "" {
					data = r.Get(item.Key)
				}
			}
			break
		default:
			data = ""
		}
		if data != "" {
			err = validate.Var(data, item.Validation)
			if err != nil {
				return res, fmt.Errorf("%s is not valid", item.Key)
			}
			res = append(res, data)
		} else if item.Default != "" {
			res = append(res, item.Default)
		} else {
			return res, fmt.Errorf("%s is empty", item.Key)
		}

	}

	if len(res) != len(parser.Params) {
		return res, fmt.Errorf("unable to parse inputs")
	}
	if parser.Processor != nil {
		parser.Processor(res)
	}
	return res, nil
}
