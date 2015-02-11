package flotilla

import (
	"log"
	"strings"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/xrr"
)

var (
	configureFirst = []Configuration{
		cengine,
	}

	configureLast = []Configuration{
		cstatic,
		cblueprints,
		ctemplating,
		csession,
	}
)

type (
	Config struct {
		Configured    bool
		Configuration []Configuration
		deferred      []Configuration
	}

	Configuration func(*App) error
)

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
	a.Configuration = append(a.Configuration, cnf...)
	err = runConf(a, a.Configuration...)
	err = runConf(a, a.Config.deferred...)
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
	a.Env.TemplatorInit()
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

func Mode(mode string, value bool) Configuration {
	return func(a *App) error {
		m := strings.Title(mode)
		if existsIn(m, []string{"Development", "Testing", "Production"}) {
			err := a.SetMode(m, value)
			if err != nil {
				return err
			}
		} else {
			return xrr.NewError("mode must be Development, Testing, or Production; not %s", mode)
		}
		return nil
	}
}

func EnvItem(items ...string) Configuration {
	return func(a *App) error {
		for _, item := range items {
			v := strings.Split(item, ":")
			k, value := v[0], v[1]
			sl := strings.Split(k, "_")
			if len(sl) > 1 {
				section, label := sl[0], sl[1]
				a.Env.Store.add(section, label, value)
			} else {
				a.Env.Store.add("", sl[0], value)
			}
		}
		return nil
	}
}

func Extension(name string, fn interface{}) Configuration {
	return func(a *App) error {
		return a.Env.AddExtension(name, fn)
	}
}

func Extensions(fns map[string]interface{}) Configuration {
	return func(a *App) error {
		return a.Env.AddExtensions(fns)
	}
}

func Templating(t Templator) Configuration {
	return func(a *App) error {
		a.Env.Templator = t
		return nil
	}
}

func TemplateFunction(name string, fn interface{}) Configuration {
	return func(a *App) error {
		a.Env.AddTplFunc(name, fn)
		return nil
	}
}

func TemplateFunctions(fns map[string]interface{}) Configuration {
	return func(a *App) error {
		a.Env.AddTplFuncs(fns)
		return nil
	}
}

func CtxProcessor(name string, fn interface{}) Configuration {
	return func(a *App) error {
		a.CtxProcessor(name, fn)
		return nil
	}
}

func CtxProcessors(fns map[string]interface{}) Configuration {
	return func(a *App) error {
		a.CtxProcessors(fns)
		return nil
	}
}

func Logger(l *log.Logger) Configuration {
	return func(a *App) error {
		a.Messaging.Logger = l
		return nil
	}
}
