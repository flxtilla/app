package app_test

import (
	"testing"

	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/txst"
)

//func AppForTest(t *testing.T, name string, conf ...Config) *App {
//	conf = append(conf, Mode("Testing", true))
//	a := New(name, conf...)
//	err := a.Configure()
//	if err != nil {
//		t.Errorf("Error in app configuration: %s", err.Error())
//	}
//	return a
//}

func TestSimple(t *testing.T) {
	a := txst.TxstingApp(t, "simple")
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

	app := txst.TxstingApp(t, "flotilla_testRouteOK")

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
	rtxs := txs()
	for _, r := range rtxs {
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
	app := txst.TxstingApp(t, "multipleRoutesSameMethodOk")
	txst.MultiPerformer(t, app, x...).Perform()
	for _, rtx := range rtxs {
		if !rtx.passed && !rtx.runonce {
			t.Errorf("Multiple route same method error: %+v", rtx)
		}
	}
}

func testRouteNotOK(method string, t *testing.T) {
	exp, _ := txst.NotFoundExpectation(
		method,
		"/test",
		func(t *testing.T) state.Manage {
			return func(s state.State) {}
		},
	)

	app := txst.TxstingApp(t, "flotilla_testRouteNotOk")

	txst.SimplePerformer(t, app, exp).Perform()
}

func TestRouteNotOK(t *testing.T) {
	for _, m := range txst.METHODS {
		testRouteNotOK(m, t)
	}
}
