package e

import "fmt"

type Error struct {
	Type     string         `json:"type"`
	Field    string         `json:"field,omitempty"`
	Message  string         `json:"message"`
	Solution string         `json:"solution,omitempty"`
	Params   *[]interface{} `json:"params,omitempty"`
}

type Errors []Error

func New(t, field, message, solution string, params ...interface{}) *Error {
	return &Error{
		t, field, message, solution, &params,
	}
}

func Field(field, message interface{}, params ...interface{}) *Error {
	return New("field", fmt.Sprint(field), fmt.Sprint(message), "", params)
}

func Context(message interface{}, params ...interface{}) *Error {
	return New("context", "", fmt.Sprint(message), "", params)
}

func (e *Error) SetSolution(s string) *Error {
	e.Solution = s
	return e
}

func (e *Error) SetParams(params ...interface{}) *Error {
	e.Params = &params
	return e
}

func (e *Error) SetType(t string) *Error {
	e.Type = t
	return e
}

func (e *Error) SetMessage(message string) *Error {
	e.Message = message
	return e
}

func (e *Error) SetFiled(field string) *Error {
	e.Field = field
	return e
}

func (e *Errors) Push(error *Error) *Errors {
	*e = append(*e, *error)
	return e
}

func (e *Errors) Exist() bool {
	return len(*e) != 0
}

func (e *Errors) Clear() *Errors {
	e = &Errors{}
	return e
}
