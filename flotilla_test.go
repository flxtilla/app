package flotilla

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

var METHODS []string = []string{"GET", "POST", "PATCH", "DELETE", "PUT", "OPTIONS", "HEAD"}

func testConf(cf ...Configuration) []Configuration {
	var ret []Configuration
	for _, c := range cf {
		ret = append(ret, c)
	}
	ret = append(ret, Mode("testing", true))
	return ret
}

func testRoutes(rs ...*Route) []*Route {
	var ret []*Route
	for _, r := range rs {
		ret = append(ret, r)
	}
	return ret
}

func testApp(t *testing.T, name string, conf []Configuration, routes []*Route) *App {
	a := New(name, conf...)
	a.Messaging.Queues["out"] = testout(t, a)
	a.Messaging.Queues["panic"] = testpanicq(t, a)
	for _, r := range routes {
		a.Manage(r)
	}
	err := a.Configure()
	if err != nil {
		t.Errorf("Error in app configuration: %s", err.Error())
	}
	return a
}

type Performer struct {
	t        *testing.T
	h        http.Handler
	code     int
	request  *http.Request
	response *httptest.ResponseRecorder
}

func NewPerformer(t *testing.T, a *App, code int, method, path string) *Performer {
	req, _ := http.NewRequest(method, path, nil)
	res := httptest.NewRecorder()

	return &Performer{
		t:        t,
		h:        a,
		code:     code,
		request:  req,
		response: res,
	}
}

func performFor(p *Performer) *Performer {
	p.h.ServeHTTP(p.response, p.request)

	if p.response.Code != p.code {
		p.t.Errorf("%s :: %s\nStatus code should be %d, was %d\n", p.request.Method, p.request.URL.Path, p.code, p.response.Code)
	}

	return p
}

func TestSimple(t *testing.T) {
	a := testApp(t, "simple", nil, nil)
	if a.Name() != "simple" {
		t.Errorf(`App name was %s, expected "simple"`, a.Name())
	}
}

func testRouteOK(method string, t *testing.T) {
	var passed bool = false

	r := NewRoute(defaultRouteConf(method, "/test", []Manage{func(c Ctx) {
		passed = true
	}}))

	f := testApp(t, "flotilla_testRouteOK", nil, testRoutes(r))

	p := NewPerformer(t, f, 200, method, "/test")

	performFor(p)

	if passed == false {
		t.Errorf("Route handler %s was not invoked.", method)
	}
}

func TestRouteOK(t *testing.T) {
	for _, m := range METHODS {
		testRouteOK(m, t)
	}
}

type rx struct {
	method  string
	passed  bool
	runonce bool
	rt      *Route
}

func TestMultipleRoutesSameMethodOK(t *testing.T) {
	var rtx []*rx
	for _, m := range METHODS {
		mkrx := &rx{
			method:  m,
			runonce: false,
			passed:  false,
		}
		mkrx.rt = NewRoute(defaultRouteConf(m, "/test", []Manage{func(c Ctx) { mkrx.passed, mkrx.runonce = true, true }}))
		rtx = append(rtx, mkrx)
	}
	var rts []*Route
	for _, x := range rtx {
		rts = append(rts, x.rt)
	}
	a := testApp(t, "testRoutesOK", nil, testRoutes(rts...))
	for _, m := range METHODS {
		p := NewPerformer(t, a, 200, m, "/test")
		performFor(p)
	}
	for _, x := range rtx {
		if x.passed != true && x.runonce != true {
			t.Errorf("Route with same path, but differing method was not registered or run: %+v", x)
		}
	}
}

func methodNotMethod(method string) string {
	newmethod := METHODS[rand.Intn(len(METHODS))]
	if newmethod == method {
		methodNotMethod(newmethod)
	}
	return newmethod
}

func testRouteNotOK(method string, t *testing.T) {
	var passed bool = false

	othermethod := methodNotMethod(method)

	r := NewRoute(defaultRouteConf(othermethod, "/test_notfound", []Manage{func(c Ctx) {
		passed = true
	}}))

	f := testApp(t, "flotilla_testRouteNotOk", nil, testRoutes(r))

	p := NewPerformer(t, f, 404, method, "/test_notfound")

	performFor(p)

	if passed == true {
		t.Errorf("Route handler %s was not invoked.", method)
	}
}

func TestRouteNotOK(t *testing.T) {
	for _, m := range METHODS {
		testRouteNotOK(m, t)
	}
}
