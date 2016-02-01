package app

import (
	"net/http"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/state"
)

type StateMakerFn func(a *App) state.Make

type Statr interface {
	StateFunction(a *App) state.Make
	SwapStateFunction(StateMakerFn)
}

type defaultStatr struct {
	fn StateMakerFn
}

func DefaultStatr() Statr {
	return &defaultStatr{
		fn: defaultStateMakerFunction,
	}
}

func defaultStateMakerFunction(a *App) state.Make {
	return func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result, m []state.Manage) state.State {
		s := state.New(a.Environment, rs, a.Environment)
		s.Reset(rq, rw, m)
		s.SessionStore, _ = a.Start(s.RWriter(), s.Request())
		s.In(s)
		return s
	}
}

func (d *defaultStatr) StateFunction(a *App) state.Make {
	return d.fn(a)
}

func (d *defaultStatr) SwapStateFunction(fn StateMakerFn) {
	d.fn = fn
}
