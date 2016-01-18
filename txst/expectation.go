package txst

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thrisp/flotilla/route"
	"github.com/thrisp/flotilla/state"
)

type Expectation interface {
	Register(*testing.T, TxstApp)
	SetPre(...func(*testing.T, *http.Request))
	SetPost(...func(*testing.T, *httptest.ResponseRecorder))
	Request() *http.Request
	Response() *httptest.ResponseRecorder
	Run(*testing.T, TxstApp)
	Tanagers(...Tanage) []Tanage
}

func NewExpectation(code int, method, path string, ts ...Tanage) (*expectation, error) {
	req, err := http.NewRequest(method, path, nil)
	exp := &expectation{
		code:     code,
		method:   method,
		path:     path,
		request:  req,
		response: httptest.NewRecorder(),
		tanagers: ts,
	}
	exp.SetPost(exp.defaultPost)
	return exp, err
}

func notPath(path string) string {
	return fmt.Sprintf("%s_not", path)
}

func NotFoundExpectation(method, path string, ts ...Tanage) (*expectation, error) {
	req, err := http.NewRequest(notMethod(method), notPath(path), nil)
	exp := &expectation{
		code:     404,
		method:   method,
		path:     path,
		request:  req,
		response: httptest.NewRecorder(),
		tanagers: ts,
	}
	exp.SetPost(exp.defaultPost)
	return exp, err
}

func NoTanage(code int, method, path string) (*expectation, error) {
	exp, err := NewExpectation(code, method, path)
	exp.preregistered = true
	return exp, err
}

type expectation struct {
	code          int
	method        string
	path          string
	request       *http.Request
	response      *httptest.ResponseRecorder
	prefn         []func(*testing.T, *http.Request)
	postfn        []func(*testing.T, *httptest.ResponseRecorder)
	preregistered bool
	tanagers      []Tanage
}

func (e *expectation) Register(t *testing.T, a TxstApp) {
	var reg []state.Manage
	for _, m := range e.tanagers {
		reg = append(reg, m(t))
	}
	if !e.preregistered {
		a.Manage(route.New(
			func(rt *route.Route) error {
				rt.Method = e.method
				rt.Base = e.path
				rt.Managers = reg
				return nil
			},
		))
	}
}

func (e *expectation) Request() *http.Request {
	return e.request
}

func (e *expectation) Response() *httptest.ResponseRecorder {
	return e.response
}

func (e *expectation) pre(t *testing.T) {
	for _, p := range e.prefn {
		p(t, e.request)
	}
}

func (e *expectation) SetPre(fn ...func(*testing.T, *http.Request)) {
	e.prefn = append(e.prefn, fn...)
}

func (e *expectation) post(t *testing.T) {
	for _, p := range e.postfn {
		p(t, e.response)
	}
}

func (e *expectation) SetPost(fn ...func(*testing.T, *httptest.ResponseRecorder)) {
	e.postfn = append(e.postfn, fn...)
}

func (e *expectation) defaultPost(t *testing.T, r *httptest.ResponseRecorder) {
	if r.Code != e.code {
		t.Errorf(
			"%s :: %s\nStatus code should be %d, was %d\n",
			e.request.Method,
			e.request.URL.Path,
			e.code,
			r.Code,
		)
	}
}

func (e *expectation) Run(t *testing.T, a TxstApp) {
	e.pre(t)
	a.ServeHTTP(e.response, e.request)
	e.post(t)
}

func (e *expectation) Tanagers(ts ...Tanage) []Tanage {
	e.tanagers = append(e.tanagers, ts...)
	return e.tanagers
}
