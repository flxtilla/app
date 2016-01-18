package app

import (
	"testing"

	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/txst"
)

func testApp(t *testing.T, name string, conf ...ConfigurationFn) *App {
	conf = append(conf, Mode("Testing", true))
	a := New(name, conf...)
	//mkTestQueues(t, a)
	err := a.Configure()
	if err != nil {
		t.Errorf("Error in app configuration: %s", err.Error())
	}
	return a
}

func TestSimple(t *testing.T) {
	a := testApp(t, "simple")
	if a.Name() != "simple" {
		t.Errorf(`App name was %s, expected "simple"`, a.Name())
	}
}

func testRouteOK(method string, t *testing.T) {
	var passed bool = false

	exp, _ := txst.NewExpectation(
		200,
		method,
		"/test",
		func(t *testing.T) state.Manage { return func(s state.State) { passed = true } },
	)

	app := testApp(t, "flotilla_testRouteOK")

	txst.SimplePerformer(t, app, exp).Perform()

	if passed == false {
		t.Errorf("Route handler %s was not invoked.", method)
	}
}

func TestRouteOK(t *testing.T) {
	for _, m := range txst.METHODS {
		testRouteOK(m, t)
	}
}

type tx struct {
	method  string
	passed  bool
	runonce bool
}

func txs() []*tx {
	var ret []*tx
	for _, m := range txst.METHODS {
		ret = append(ret, &tx{method: m})
	}
	return ret
}

func rmanage(t *testing.T, x *tx) state.Manage {
	return func(s state.State) {
		req := s.Request()
		if req.Method == x.method {
			x.passed = true
			x.runonce = !x.runonce
		}
		if req.Method != x.method {
			t.Errorf("Request was %s, but Manager expected %s", req.Method, x.method)
		}
	}
}

func TestMultipleRoutesSameMethodOK(t *testing.T) {
	var x []txst.Expectation
	ctxs := txs()
	for _, r := range ctxs {
		m := rmanage(t, r)
		nx, _ := txst.NewExpectation(
			200,
			r.method,
			"/test",
			func(t *testing.T) state.Manage {
				return m
			},
		)
		x = append(x, nx)
	}
	app := testApp(t, "multipleRoutesSameMethodOk")
	txst.MultiPerformer(t, app, x...).Perform()
	for _, ctx := range ctxs {
		if !ctx.passed && !ctx.runonce {
			t.Errorf("Multiple route same method error: %+v", ctx)
		}
	}
}

/*
func testRouteNotOK(method string, t *testing.T) {
	exp, _ := txst.NotFoundExpectation(
		method,
		"/test",
		func(t *testing.T) state.Manage { return func(s state.State) {} },
	)

	app := testApp(t, "flotilla_testRouteNotOk")

	txst.SimplePerformer(t, app, exp).Perform()
}

func TestRouteNotOK(t *testing.T) {
	for _, m := range txst.METHODS {
		testRouteNotOK(m, t)
	}
}
*/
