package txst

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type Performer interface {
	Register(*testing.T, TxstApp, ...Expectation)
	Perform()
	Expectations(...Expectation) []Expectation
}

// ZeroExpectationPerformer is a Performer using no Expectations.
func ZeroExpectationPerformer(t *testing.T, a TxstApp, code int, method, path string) Performer {
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
	a          TxstApp
	code       int
	request    *http.Request
	response   *httptest.ResponseRecorder
	registered bool
}

func (p *zeroExpectationPerformer) Register(t *testing.T, a TxstApp, es ...Expectation) {
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
func SimplePerformer(t *testing.T, a TxstApp, es ...Expectation) Performer {
	p := &simplePerformer{}
	p.Register(t, a, es...)
	return p
}

type simplePerformer struct {
	t          *testing.T
	a          TxstApp
	e          Expectation
	registered bool
}

func (p *simplePerformer) Register(t *testing.T, a TxstApp, es ...Expectation) {
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
func MultiPerformer(t *testing.T, a TxstApp, es ...Expectation) Performer {
	p := &multiPerformer{}
	p.Register(t, a, es...)
	return p
}

type multiPerformer struct {
	t          *testing.T
	a          TxstApp
	e          []Expectation
	registered bool
}

func (p *multiPerformer) Register(t *testing.T, a TxstApp, e ...Expectation) {
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
func SessionPerformer(t *testing.T, a TxstApp, es ...Expectation) Performer {
	p := &sessionPerformer{cj: make(map[string]*http.Cookie)}
	p.Register(t, a, es...)
	return p
}

type sessionPerformer struct {
	t          *testing.T
	a          TxstApp
	e          []Expectation
	cj         map[string]*http.Cookie
	registered bool
}

func (p *sessionPerformer) Register(t *testing.T, a TxstApp, e ...Expectation) {
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
