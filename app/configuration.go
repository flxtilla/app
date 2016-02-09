package app

import (
	"strings"

	"github.com/thrisp/flotilla/asset"
	"github.com/thrisp/flotilla/blueprint"
	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/extension"
)

// A ConfigurationFn is any function taking an App instance and returning an
// error.
type ConfigurationFn func(*App) error

var configureImmediate = []ConfigurationFn{
	cEnsureEnvironment,
	cEnsureEngine,
	cEnsureBlueprints,
}

var configureLast = []ConfigurationFn{
	cRegisterBlueprints,
	cSession,
	cRegisterTemplateRender,
}

// Configuration is an interface to App configuration.
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

// UseFn adds a ConfigurationFn to the Configuration to one of three queues
// specified by a string. The default Configuration uses preferrred, referred,
// and deferred queues for functions. These are useful if you need to ensure a
// configuration function runs before or after other configuration functions.
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

// Configure runs all actions for the configuration instance.
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

// Configured returns a boolean indicating if Configuration has been run.
func (c *configuration) Configured() bool {
	return c.configured
}

func cEnsureEnvironment(a *App) error {
	if a.Environment == nil {
		a.Environment = newEnvironment(a)
	}
	return nil
}

func cEnsureEngine(a *App) error {
	if a.Engine == nil {
		a.Engine = engine.DefaultEngine(nil)
	}
	return nil
}

func cEnsureBlueprints(a *App) error {
	if a.Blueprints == nil {
		a.Blueprints = blueprint.NewBlueprints("/", a.Handle, a.StateFunction(a))
	}
	a.Handle("STATUS", "DEFAULT", a.StatusRule())
	return nil
}

func cSession(a *App) error {
	a.Init()
	return nil
}

func cRegisterBlueprints(a *App) error {
	for _, b := range a.ListBlueprints() {
		if !b.Registered() {
			b.Register()
		}
	}
	return nil
}

func cRegisterTemplateRender(a *App) error {
	a.SetTemplateFunctions()
	ext := extension.New(
		"template_render_extension",
		mkFunction("render_template", a.RenderTemplate),
	)
	a.Extend(ext)
	return nil
}

// Mode returns a ConfigurationFn for the mode and value, e.g. Mode("testing",
// true).
func Mode(mode string, value bool) ConfigurationFn {
	return func(a *App) error {
		return a.SetMode(mode, value)
	}
}

// Store returns a ConfigurationFn that adds key value items to the environment
// Store in the form of "key:value".
func Store(items ...string) ConfigurationFn {
	return func(a *App) error {
		for _, item := range items {
			v := strings.Split(item, ":")
			key, value := v[0], v[1]
			a.Add(key, value)
		}
		return nil
	}
}

// Extend returns a ConfigurationFn that adds the provided extension.Extensions
// to the app Environment.
func Extend(fxs ...extension.Extension) ConfigurationFn {
	return func(a *App) error {
		a.Extend(fxs...)
		return nil
	}
}

// Assets returns a ConfigurationFn adding the provided AssetFS to the app
// Environment Assets.
func Assets(as ...asset.AssetFS) ConfigurationFn {
	return func(a *App) error {
		a.SetAssetFS(as...)
		return nil
	}
}
