package flotilla

import (
	"strings"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/xrr"
)

var configureFirst = []Configuration{
	cengine,
}

var configureLast = []Configuration{
	cstatic,
	cblueprints,
	ctemplating,
	csession,
}

type Config struct {
	Configured    bool
	Configuration []Configuration
	deferred      []Configuration
}

type Configuration func(*App) error

func newConfig(cnf ...Configuration) *Config {
	return &Config{
		Configured:    false,
		Configuration: cnf,
		deferred:      configureLast,
	}
}

func runConf(a *App, cnf ...Configuration) error {
	var err error
	for _, fn := range cnf {
		err = fn(a)
	}
	return err
}

func (a *App) Configure(cnf ...Configuration) error {
	var err error
	configuration := append(a.Configuration, cnf...)
	err = runConf(a, configuration...)
	if err != nil {
		return err
	}
	deferred := a.Config.deferred
	err = runConf(a, deferred...)
	if err != nil {
		return err
	}
	a.Configured = true
	return nil
}

func cengine(a *App) error {
	if a.Engine == nil {
		a.Engine = engine.DefaultEngine(StatusRule(a))
	}
	return nil
}

func csession(a *App) error {
	a.Env.SessionInit()
	return nil
}

func ctemplating(a *App) error {
	TemplatorInit(a)
	return nil
}

func cstatic(a *App) error {
	StaticorInit(a)
	return nil
}

func cblueprints(a *App) error {
	for _, b := range a.Blueprints() {
		if !b.registered {
			b.Register(a)
		}
	}
	return nil
}

var IllegalMode = xrr.NewXrror("mode must be Development, Testing, or Production; not %s").Out

func Mode(mode string, value bool) Configuration {
	return func(a *App) error {
		m := strings.Title(mode)
		if existsIn(m, []string{"Development", "Testing", "Production"}) {
			err := a.SetMode(m, value)
			if err != nil {
				return err
			}
			return nil
		}
		return IllegalMode(mode)
	}
}

func EnvItem(items ...string) Configuration {
	return func(a *App) error {
		for _, item := range items {
			v := strings.Split(item, ":")
			k, value := v[0], v[1]
			sl := strings.Split(k, "_")
			if len(sl) > 1 {
				section, label := sl[0], strings.Join(sl[1:], "_")
				a.Env.Store.add(section, label, value)
			} else {
				a.Env.Store.add("", sl[0], value)
			}
		}
		return nil
	}
}

func Extensions(fxs ...Fxtension) Configuration {
	return func(a *App) error {
		return a.Env.AddFxtensions(fxs...)
	}
}

func UseStaticor(s Staticor) Configuration {
	return func(a *App) error {
		a.Env.Staticor = s
		return nil
	}
}

func UseTemplator(t Templator) Configuration {
	return func(a *App) error {
		a.Env.Templator = t
		return nil
	}
}

func CtxProcessor(name string, fn interface{}) Configuration {
	return func(a *App) error {
		a.AddCtxProcessor(name, fn)
		return nil
	}
}

func CtxProcessors(fns map[string]interface{}) Configuration {
	return func(a *App) error {
		a.AddCtxProcessors(fns)
		return nil
	}
}

func WithAssets(ast ...AssetFS) Configuration {
	return func(a *App) error {
		a.Env.Assets = append(a.Env.Assets, ast...)
		return nil
	}
}

func WithQueue(name string, q Queue) Configuration {
	return func(a *App) error {
		a.Messaging.Queues[name] = q
		return nil
	}
}
