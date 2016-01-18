package state

import (
	"net/http"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/extension"
	"github.com/thrisp/flotilla/flash"
	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/xrr"
)

type Make func(
	http.ResponseWriter,
	*http.Request,
	*engine.Result,
	[]Manage,
) State

type Manage func(State)

type State interface {
	flash.Flasher
	extension.Fxtension
	xrr.Xrroror
	session.SessionStore
	Handlers
	Request() *http.Request
	RWriter() ResponseWriter
	Reset(*http.Request, http.ResponseWriter, []Manage)
	Replicate() State
	Run()
	Rerun(...Manage)
	Next()
	Cancel()
}

type Handlers interface {
	Push(Manage)
	Bounce(Manage)
}

type handlers struct {
	index    int8
	managers []Manage
	deferred []Manage
}

func defaultHandlers() *handlers {
	return &handlers{index: -1}
}

func (h *handlers) Push(fn Manage) {
	h.deferred = append(h.deferred, fn)
}

func (h *handlers) Bounce(fn Manage) {
	h.deferred = []Manage{fn}
}

type state struct {
	*engine.Result
	//*context
	*handlers
	xrr.Xrroror
	extension.Fxtension
	session.SessionStore
	rw      responseWriter
	RW      ResponseWriter
	request *http.Request
	Data    map[string]interface{}
	flash.Flasher
}

func empty() *state {
	return &state{
		handlers: defaultHandlers(),
		Xrroror:  xrr.NewXrroror(),
	}
}

func New(ext extension.Fxtension, rs *engine.Result) *state {
	s := empty()
	s.Result = rs
	s.Fxtension = ext
	s.RW = &s.rw
	s.Flasher = flash.New()
	return s
}

func (s *state) Request() *http.Request {
	return s.request
}

func (s *state) RWriter() ResponseWriter {
	return s.RW
}

func release(s State) {
	w := s.RWriter()
	if !w.Written() {
		s.Out(s)
		s.SessionRelease(w)
	}
}

func (s *state) Run() {
	s.Push(release)
	s.Next()
	for _, fn := range s.deferred {
		fn(s)
	}
	//if !ModeIs(s, "production") {
	s.PostProcess(s.request, s.RW.Status())
	//s.Call("out", LogFmt(s))
	//}
}

func (s *state) Rerun(managers ...Manage) {
	if s.index != -1 {
		s.index = -1
	}
	s.managers = managers
	s.Next()
}

func (s *state) Next() {
	s.index++
	lm := int8(len(s.managers))
	for ; s.index < lm; s.index++ {
		s.managers[s.index](s)
	}
}

func (s *state) Cancel() {
	s.PostProcess(s.request, s.RW.Status())
	//c.context.cancel(true, Canceled)
}

func (s *state) Reset(rq *http.Request, rw http.ResponseWriter, m []Manage) {
	s.request = rq
	s.rw.reset(rw)
	//c.context = &context{done: make(chan struct{}), value: c}
	s.handlers = defaultHandlers()
	s.managers = m
}

func (s *state) Replicate() State {
	//child := &context{parent: c.context, done: make(chan struct{}), value: c}
	//propagateCancel(c.context, child)
	var rcopy state = *s
	//rcopy.context = child
	return &rcopy
}
