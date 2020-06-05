package constant

import (
	"github.com/pkg/errors"
)

var (
	ERROR_JSON_PARSE       = errors.New("unable to parse json")
	ERROR_FORM_PARSE       = errors.New("unable to parse form")
	ERROR_INVALID_ID       = errors.New("invalid id")
	ERROR_OBJECT_NOT_EXIST = errors.New("object does not exist")
	ERROR_UNAUTHORIZED     = errors.New("unauthorized access")
	ERROR_MUST_LOGIN       = errors.New("you must login first")
)
