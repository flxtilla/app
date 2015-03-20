package flotilla

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func callStatus(status int) Manage {
	return func(c Ctx) {
		_, _ = c.Call("status", status)
	}
}

func callStatusRoute(method, route string, status int) *Route {
	return NewRoute(method, route, false, []Manage{callStatus(status)})
}

func routestring(status int) string {
	return fmt.Sprintf("/call%d", status)
}

func testStatus(method, expects string, status int, t *testing.T) {
	rts := testRoutes(callStatusRoute(method, routestring(status), status))
	a := testApp(t, "statuses", nil, rts)

	p := NewPerformer(t, a, status, method, routestring(status))

	performFor(p)

	if bytes.Compare(p.response.Body.Bytes(), []byte(expects)) != 0 {
		t.Errorf("Status test expected %s, but got %s.", expects, p.response.Body.String())
	}
}

func TestStatus(t *testing.T) {
	for _, m := range METHODS {
		testStatus(m, "418 I'm a teapot", 418, t)
	}
}

func testpanic(c Ctx) {
	panic("Test panic!")
}

func test500(method string, t *testing.T) {
	pnc := NewRoute(method, routestring(500), false, []Manage{testpanic})
	a := testApp(t, "panic", nil, testRoutes(pnc))
	p := NewPerformer(t, a, 500, method, routestring(500))
	performFor(p)
	if !strings.Contains(p.response.Body.String(), "Test panic!") {
		t.Errorf(`Status test 500 expected to contain "Test Panic!", but did not.`)
	}
}

func Test500(t *testing.T) {
	for _, m := range METHODS {
		test500(m, t)
	}
}

func Custom404(c Ctx) {
	c.Call("serveplain", 404, "I AM NOT FOUND :: 404")
}

func customStatus404(method, expects string, t *testing.T) {
	a := testApp(t, "customstatus404", nil, nil)

	a.STATUS(404, Custom404)

	p := NewPerformer(t, a, 404, method, "/notfound")

	performFor(p)

	if bytes.Compare(p.response.Body.Bytes(), []byte(expects)) != 0 {
		t.Errorf("Custom status 404 test expected %s, but got %s.", expects, p.response.Body.String())
	}
}

func Custom418(c Ctx) {
	c.Call("serveplain", 418, "I AM TEAPOT :: 418")
}

func customStatus(method, expects string, status int, m Manage, t *testing.T) {
	rts := testRoutes(callStatusRoute(method, routestring(status), status))
	a := testApp(t, "customstatus", nil, rts)

	a.STATUS(status, m)

	p := NewPerformer(t, a, status, method, routestring(status))

	performFor(p)

	if bytes.Compare(p.response.Body.Bytes(), []byte(expects)) != 0 {
		t.Errorf("Custom status %d test expected %s, but got %s.", status, expects, p.response.Body.String())
	}
}

func TestCustomStatus(t *testing.T) {
	for _, m := range METHODS {
		customStatus404(m, "I AM NOT FOUND :: 404", t)
		customStatus(m, "I AM TEAPOT :: 418", 418, Custom418, t)
	}
}
