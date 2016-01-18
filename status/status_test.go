package status

/*
import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func callStatus(status int) Manage {
	return func(c Ctx) {
		_, _ = c.Call("status", status)
	}
}

func routestring(status int) string {
	return fmt.Sprintf("/call%d", status)
}

func testStatus(t *testing.T, status int, method, expects string) {
	a := testApp(t, "statuses")

	exp, _ := NewExpectation(
		status,
		method,
		routestring(status),
		func(t *testing.T) Manage {
			return callStatus(status)
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			if bytes.Compare(r.Body.Bytes(), []byte(expects)) != 0 {
				t.Errorf("Status test expected %s, but got %s.", expects, r.Body.String())
			}
		},
	)

	SimplePerformer(t, a, exp).Perform()
}

func TestStatus(t *testing.T) {
	for _, m := range METHODS {
		testStatus(t, 418, m, "418 I'm a teapot")
	}
}

func test500(method string, t *testing.T) {
	a := testApp(t, "panic")
	exp, _ := NewExpectation(
		500,
		method,
		routestring(500),
		func(t *testing.T) Manage {
			return func(c Ctx) {
				panic("Test panic!")
			}
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			if !strings.Contains(r.Body.String(), "Test panic!") {
				t.Errorf(`Status test 500 expected to contain "Test Panic!", but did not.`)
			}
		},
	)
	SimplePerformer(t, a, exp).Perform()
}

func Test500(t *testing.T) {
	for _, m := range METHODS {
		test500(m, t)
	}
}

func Custom404(c Ctx) {
	c.Call("serveplain", 404, "I AM NOT FOUND :: 404")
}

func customStatus404(t *testing.T, method, expects string) {
	a := testApp(t, "customstatus404")
	a.STATUS(404, Custom404)

	z := ZeroExpectationPerformer(t, a, 404, method, "/notfound")
	z.Perform()

	if bytes.Compare(z.response.Body.Bytes(), []byte(expects)) != 0 {
		t.Errorf("Custom status 404 test expected %s, but got %s.", expects, z.response.Body.String())
	}
}

func Custom418(c Ctx) {
	c.Call("serveplain", 418, "I AM TEAPOT :: 418")
}

func customStatus(t *testing.T, method string, expects string, status int, m Manage) {
	a := testApp(t, "customstatus")
	a.STATUS(status, m)

	exp, _ := NewExpectation(
		418,
		"GET",
		routestring(status),
		func(t *testing.T) Manage {
			return func(c Ctx) {
				_, _ = c.Call("status", status)
			}
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			if bytes.Compare(r.Body.Bytes(), []byte(expects)) != 0 {
				t.Errorf("Custom status %d test expected %s, but got %s.", status, expects, r.Body.String())
			}
		},
	)

	SimplePerformer(t, a, exp).Perform()
}

func TestCustomStatus(t *testing.T) {
	for _, m := range METHODS {
		customStatus404(t, m, "I AM NOT FOUND :: 404")
		customStatus(t, m, "I AM TEAPOT :: 418", 418, Custom418)
	}
}
*/
