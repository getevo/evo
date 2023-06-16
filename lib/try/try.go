package try

import "github.com/getevo/evo/v2/lib/panics"

const rethrow_panic = "_____rethrow"

type (
	Error     interface{}
	exception struct {
		finally   func()
		Error     Error
		Recovered *panics.Recovered
	}
)

func Throw() {
	panic(rethrow_panic)
}

func This(f func()) (e exception) {
	e = exception{nil, nil, nil}
	// catch error in
	var pc panics.Catcher
	pc.Try(f)
	recovered := pc.Recovered()
	e.Error = recovered.AsError()
	e.Recovered = pc.Recovered()
	return
}

func (e exception) Catch(f func(err *panics.Recovered)) {
	if e.Error != nil {
		f(e.Recovered)
	} else if e.finally != nil {
		e.finally()
	}
}

func (e exception) Finally(f func()) (e2 exception) {
	if e.finally != nil {
		panic("finally was only set")
	}
	e2 = e
	e2.finally = f
	return
}
