package flotilla

import (
	"net/http"
	"testing"

	"github.com/thrisp/flotilla/engine"
)

type tc struct {
	h Manage
}

func (c *tc) Run() {
	c.Next()
}

func (c *tc) Next() {
	c.h(c)
}

func (c *tc) Call(name string, vals ...interface{}) (interface{}, error) {
	return nil, nil
}

func (c *tc) Cancel() {}

func MakeTestCtx(rw http.ResponseWriter, rq *http.Request, rs *engine.Result, rt *Route) Ctx {
	c := &tc{h: rt.Managers[0]}
	return c
}

func testctx(method string, t *testing.T) {
	var passed bool = false

	r := NewRoute(defaultRouteConf(method, "/test_ctx", []Manage{func(c Ctx) { passed = true }}))

	a := Base("test_ctx")

	b := NewBlueprint("/")

	b.MakeCtx = MakeTestCtx

	a.Blueprint = b

	a.Config = newConfig()

	a.Manage(r)

	a.Configure()

	ZeroExpectationPerformer(t, a, 200, method, "/test_ctx").Perform()

	if passed == false {
		t.Errorf("Test Ctx route handler %s was not invoked.", method)
	}
}

func TestCtx(t *testing.T) {
	for _, m := range METHODS {
		testctx(m, t)
	}
}
