package flotilla

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var METHODS []string = []string{"GET", "POST", "PATCH", "DELETE", "PUT", "OPTIONS", "HEAD"}

func notMethod(method string) string {
	newmethod := METHODS[rand.Intn(len(METHODS))]
	if newmethod == method {
		notMethod(method)
	}
	return newmethod
}

type Tanage func(*testing.T) Manage

type Expectation interface {
	Register(*testing.T, *App)
	SetPre(...func(*testing.T, *http.Request))
	SetPost(...func(*testing.T, *httptest.ResponseRecorder))
	Request() *http.Request
	Response() *httptest.ResponseRecorder
	Run(*testing.T, *App)
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

func (e *expectation) Register(t *testing.T, a *App) {
	var reg []Manage
	for _, m := range e.tanagers {
		reg = append(reg, m(t))
	}
	if !e.preregistered {
		a.Manage(NewRoute(
			func(rt *Route) error {
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

func (e *expectation) Run(t *testing.T, a *App) {
	e.pre(t)
	a.ServeHTTP(e.response, e.request)
	e.post(t)
}

func (e *expectation) Tanagers(ts ...Tanage) []Tanage {
	e.tanagers = append(e.tanagers, ts...)
	return e.tanagers
}

type Performer interface {
	Register(*testing.T, *App, ...Expectation)
	Perform()
	Expectations(...Expectation) []Expectation
}

// ZeroExpectationPerformer is a Performer using no Expectations.
func ZeroExpectationPerformer(t *testing.T, a *App, code int, method, path string) *zeroExpectationPerformer {
	req, _ := http.NewRequest(method, path, nil)
	res := httptest.NewRecorder()
	p := &zeroExpectationPerformer{
		code:     code,
		request:  req,
		response: res,
	}
	p.Register(t, a)
	return p
}

type zeroExpectationPerformer struct {
	t          *testing.T
	a          *App
	code       int
	request    *http.Request
	response   *httptest.ResponseRecorder
	registered bool
}

func (p *zeroExpectationPerformer) Register(t *testing.T, a *App, es ...Expectation) {
	p.t = t
	p.a = a
	p.registered = true
}

func (p *zeroExpectationPerformer) Perform() {
	if !p.registered {
		p.t.Errorf("*zeroExpectationPerformer is not registered or properly configured: %#+v", p)
	}

	p.a.ServeHTTP(p.response, p.request)

	if p.response.Code != p.code {
		p.t.Errorf(
			"\n%s :: %s :: Status code should be %d, was %d\n",
			p.request.Method,
			p.request.URL.Path,
			p.code,
			p.response.Code,
		)
	}

}

func (p *zeroExpectationPerformer) Expectations(es ...Expectation) []Expectation {
	return []Expectation{}
}

// SimplePerformer is a Performer for a maximum of one Expectation.
func SimplePerformer(t *testing.T, a *App, es ...Expectation) Performer {
	p := &simplePerformer{}
	p.Register(t, a, es...)
	return p
}

type simplePerformer struct {
	t          *testing.T
	a          *App
	e          Expectation
	registered bool
}

func (p *simplePerformer) Register(t *testing.T, a *App, es ...Expectation) {
	p.t = t
	p.a = a
	p.e = es[0]
	p.e.Register(p.t, p.a)
	p.registered = true
}

func (p *simplePerformer) Perform() {
	if !p.registered {
		p.t.Errorf("*simpleperformer is not registered or properly configured: %#+v", p)
	}
	p.e.Run(p.t, p.a)
}

func (p *simplePerformer) Expectations(es ...Expectation) []Expectation {
	p.e = es[0]
	return []Expectation{p.e}
}

// MultiPerformer is a Performer for any number of Expectations.
func MultiPerformer(t *testing.T, a *App, es ...Expectation) Performer {
	p := &multiPerformer{}
	p.Register(t, a, es...)
	return p
}

type multiPerformer struct {
	t          *testing.T
	a          *App
	e          []Expectation
	registered bool
}

func (p *multiPerformer) Register(t *testing.T, a *App, e ...Expectation) {
	p.t = t
	p.a = a
	p.e = e
	for _, e := range p.e {
		e.Register(p.t, p.a)
	}
	p.registered = true
}

func (p *multiPerformer) Perform() {
	if !p.registered {
		p.t.Errorf("*multiPerformer is not registered or properly configured: %#+v", p)
	}
	for _, e := range p.e {
		e.Run(p.t, p.a)
	}
}

func (p *multiPerformer) Expectations(es ...Expectation) []Expectation {
	p.e = append(p.e, es...)
	return p.e
}

// SessionPerformer is a Performer for any number of Expectations using the same session.
func SessionPerformer(t *testing.T, a *App, es ...Expectation) Performer {
	p := &sessionPerformer{cj: make(map[string]*http.Cookie)}
	p.Register(t, a, es...)
	return p
}

type sessionPerformer struct {
	t          *testing.T
	a          *App
	e          []Expectation
	cj         map[string]*http.Cookie
	registered bool
}

func (p *sessionPerformer) Register(t *testing.T, a *App, e ...Expectation) {
	p.t = t
	p.a = a
	p.e = e
	for _, e := range p.e {
		e.Register(p.t, p.a)
	}
	p.registered = true
}

func extractCookie(ck string) *http.Cookie {
	var ret = &http.Cookie{}
	ret.Raw = ck
	sck := strings.Split(ck, "; ")
	ret.Unparsed = sck
	nv := strings.SplitN(sck[0], "=", 2)
	ret.Name, ret.Value = nv[0], nv[1]
	for _, f := range sck[0:] {
		s := strings.SplitN(f, "=", 2)
		if len(s) > 1 {
			k, v := s[0], s[1]
			switch {
			case k == "Max-Age":
				mai, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					mai = 0
				}
				ret.MaxAge = int(mai)
			case k == "Path":
				ret.Path = v
			case k == "Domain":
				ret.Domain = v
			case k == "Expires":
				ret.RawExpires = v
			}
		}
		if len(s) == 1 {
			switch {
			case s[0] == "HttpOnly":
				ret.HttpOnly = true
			case s[0] == "Secure":
				ret.Secure = true
			}
		}
	}
	return ret
}

func extractCookies(r *httptest.ResponseRecorder) []*http.Cookie {
	var ret []*http.Cookie
	cks := r.HeaderMap["Set-Cookie"]
	for _, ck := range cks {
		ret = append(ret, extractCookie(ck))
	}
	return ret
}

func addCookies(r *http.Request, cks map[string]*http.Cookie) {
	for _, ck := range cks {
		r.AddCookie(ck)
	}
}

func (p *sessionPerformer) updateCookies(cks []*http.Cookie) {
	for _, ck := range cks {
		p.cj[ck.Name] = ck
	}
}

func (p *sessionPerformer) Perform() {
	if !p.registered {
		p.t.Errorf("*sessionPerformer is not registered or properly configured: %#+v", p)
	}
	for _, e := range p.e {
		addCookies(e.Request(), p.cj)
		e.Run(p.t, p.a)
		p.updateCookies(extractCookies(e.Response()))
	}
}

func (p *sessionPerformer) Expectations(es ...Expectation) []Expectation {
	p.e = append(p.e, es...)
	return p.e
}
