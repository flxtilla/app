package flotilla

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

var METHODS []string = []string{"GET", "POST", "PATCH", "DELETE", "PUT", "OPTIONS", "HEAD"}

func testApp(name string, routes ...*Route) *App {
	f := New(name, Mode("testing", true))
	for _, r := range routes {
		f.Manage(r)
	}
	f.Configure()
	return f
}

func TestSimple(t *testing.T) {
	a := testApp("simple")
	if a.Name() != "simple" {
		t.Errorf(`App name was %s, expected "simple"`, a.Name())
	}
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
		p.t.Errorf("Status code should be %d, was %d", p.code, p.response.Code)
	}

	return p
}

func testRouteOK(method string, t *testing.T) {
	var passed bool = false

	r := NewRoute(method, "/test", false, []Manage{func(c Ctx) {
		passed = true
	}})

	f := testApp("flotilla_testRouteOK", r)

	p := NewPerformer(t, f, 200, method, "/test")

	performFor(p)

	if passed == false {
		t.Errorf(method + " route handler was not invoked.")
	}
}

func TestRouteOK(t *testing.T) {
	for _, m := range METHODS {
		testRouteOK(m, t)
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

	r := NewRoute(othermethod, "/test_notfound", false, []Manage{func(c Ctx) {
		passed = true
	}})

	f := testApp("flotilla_testRouteNotOk", r)

	p := NewPerformer(t, f, 404, method, "/test_notfound")

	performFor(p)

	if passed == true {
		t.Errorf(method + " route handler was invoked, when it should not")
	}
}

func TestRouteNotOK(t *testing.T) {
	for _, m := range METHODS {
		testRouteNotOK(m, t)
	}
}
