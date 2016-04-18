package app

import (
	"net/http"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/state"
)

// A StateMakerFn is a function type taking an App instance, and providing a
// state.Make function.
type StateMakerFn func(a *App) state.Make

// Statr is an interface providing a StateMakerFn to the app in addition to
// functionality to change this function as needed.
type Statr interface {
	StateFunction(a *App) state.Make
	SwapStateFunction(StateMakerFn)
}

type defaultStatr struct {
	fn StateMakerFn
}

// The default flotilla Statr.
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

// The default statr StateFunction taking an App instance and returning a
// state.Make function.
func (d *defaultStatr) StateFunction(a *App) state.Make {
	return d.fn(a)
}

// The default statr SwapStateFunction for changing the statr StateMaker
// function.
func (d *defaultStatr) SwapStateFunction(fn StateMakerFn) {
	d.fn = fn
}
