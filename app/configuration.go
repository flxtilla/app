package app

import (
	"net/http"
	"strings"

	"github.com/thrisp/flotilla/blueprint"
	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/extension"
	"github.com/thrisp/flotilla/state"
	//"github.com/thrisp/flotilla/xrr"
)

type ConfigurationFn func(*App) error

var configureImmediate = []ConfigurationFn{
	cEnsureEngine,
	cEnsureEnvironment,
	cEnsureBlueprints,
}

var configureLast = []ConfigurationFn{
	cRegisterBlueprints,
	cSession,
}

type Configuration interface {
	UseFn(string, ...ConfigurationFn)
	Configure() error
	Configured() bool
}

type configuration struct {
	*App
	configured bool
	preferred  []ConfigurationFn
	referred   []ConfigurationFn
	deferred   []ConfigurationFn
}

func defaultConfiguration(a *App, cnf ...ConfigurationFn) Configuration {
	c := &configuration{
		App:        a,
		configured: false,
		preferred:  make([]ConfigurationFn, 0),
		referred:   make([]ConfigurationFn, 0),
		deferred:   make([]ConfigurationFn, 0),
	}
	c.UseFn("", cnf...)
	c.UseFn("defer", configureLast...)
	return c
}

func (c *configuration) UseFn(to string, cnf ...ConfigurationFn) {
	switch to {
	case "prefer", "preferred":
		c.preferred = append(c.preferred, cnf...)
	case "defer", "deferred":
		c.deferred = append(c.deferred, cnf...)
	default:
		c.referred = append(c.referred, cnf...)
	}
}

func runConfigure(a *App, cnf ...ConfigurationFn) error {
	var err error
	for _, fn := range cnf {
		err = fn(a)
	}
	return err
}

func (c *configuration) Configure() error {
	var run = [][]ConfigurationFn{
		c.preferred,
		c.referred,
		c.deferred,
	}

	for _, r := range run {
		err := runConfigure(c.App, r...)
		if err != nil {
			return err
		}
	}

	c.configured = true
	return nil
}

func (c *configuration) Configured() bool {
	return c.configured
}

func StatusRule(a *App) engine.Rule {
	return func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
		st := a.GetStatus(rs.Code)
		s := state.New(a.Environment, rs)
		s.Reset(rq, rw, st.Managers())
		s.Run()
		s.Cancel()
	}
}

func cEnsureEngine(a *App) error {
	if a.Engine == nil {
		a.Engine = engine.DefaultEngine(StatusRule(a))
	}
	return nil
}

func cEnsureEnvironment(a *App) error {
	if a.Environment == nil {
		a.Environment = newEnvironment(a)
	}
	return nil
}

func DefaultMakeState(a *App) state.Make {
	return func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result, m []state.Manage) state.State {
		s := state.New(a.Environment, rs)
		s.Reset(rq, rw, m)
		s.SessionStore, _ = a.Start(s.RWriter(), s.Request())
		s.In(s)
		return s
	}
}

func cEnsureBlueprints(a *App) error {
	if a.Blueprints == nil {
		a.Blueprints = blueprint.NewBlueprints("/", a.Handle, DefaultMakeState(a))
	}
	return nil
}

func cSession(a *App) error {
	a.Init()
	return nil
}

//func ctemplating(a *App) error {
//TemplatorInit(a)
//return nil
//}

//func cstatic(a *App) error {
//StaticorInit(a)
//return nil
//}

func cRegisterBlueprints(a *App) error {
	for _, b := range a.ListBlueprints() {
		if !b.Registered() {
			b.Register()
		}
	}
	return nil
}

func Mode(mode string, value bool) ConfigurationFn {
	return func(a *App) error {
		return a.SetMode(mode, value)
	}
}

func EnvItem(items ...string) ConfigurationFn {
	return func(a *App) error {
		for _, item := range items {
			v := strings.Split(item, ":")
			key, value := v[0], v[1]
			a.Add(key, value)
		}
		return nil
	}
}

func Extensions(fxs ...extension.Fxtension) ConfigurationFn {
	return func(a *App) error {
		a.Extend(fxs...)
		return nil
	}
}

//func UseStaticor(s Staticor) ConfigurationFn {
//	return func(a *App) error {
//		a.Env.Staticor = s
//		return nil
//	}
//}

//func UseTemplator(t Templator) ConfigurationFn {
//	return func(a *App) error {
//		a.Env.Templator = t
//		return nil
//	}
//}

//func CtxProcessor(name string, fn interface{}) ConfigurationFn {
//	return func(a *App) error {
//		//a.AddCtxProcessor(name, fn)
//		return nil
//	}
//}

//func CtxProcessors(fns map[string]interface{}) ConfigurationFn {
//	return func(a *App) error {
//		a.AddCtxProcessors(fns)
//		return nil
//	}
//}

//func WithAssets(ast ...asset.AssetFS) ConfigurationFn {
//	return func(a *App) error {
//		a.Env.Assets = append(a.Env.Assets, ast...)
//		return nil
//	}
//}

//func WithQueue(name string, q Queue) ConfigurationFn {
//	return func(a *App) error {
//		a.Messaging.Queues[name] = q
//		return nil
//	}
//}
