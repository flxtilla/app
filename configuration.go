package app

import (
	"log"
	"sort"
	"strings"

	"github.com/flxtilla/cxre/asset"
	"github.com/flxtilla/cxre/blueprint"
	"github.com/flxtilla/cxre/engine"
	"github.com/flxtilla/cxre/extension"
)

type ConfigFn func(*App) error

type Config interface {
	Order() int
	Configure(*App) error
}

type config struct {
	order int
	fn    ConfigFn
}

func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

func (c config) Order() int {
	return c.order
}

func (c config) Configure(a *App) error {
	return c.fn(a)
}

type configList []Config

func (c configList) Len() int {
	return len(c)
}

func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

type Configuration interface {
	AddConfig(...Config)
	AddFn(...ConfigFn)
	Configure() error
	Configured() bool
}

type configuration struct {
	a          *App
	configured bool
	list       configList
}

func newConfiguration(a *App, conf ...Config) *configuration {
	c := &configuration{
		a:    a,
		list: builtIns,
	}
	c.AddConfig(conf...)
	return c
}

func (c *configuration) AddConfig(conf ...Config) {
	c.list = append(c.list, conf...)
}

func (c *configuration) AddFn(fns ...ConfigFn) {
	for _, fn := range fns {
		c.list = append(c.list, DefaultConfig(fn))
	}
}

func configure(a *App, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(a)
		if err != nil {
			return err
		}
	}
	return nil
}

func respondTo(c *configuration, err error) {
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.a, c.list...)
	respondTo(c, err)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var preConfig = []Config{
	config{1, cEnsureEnvironment},
	config{2, cEnsureEngine},
	config{3, cEnsureBlueprints},
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

var builtIns = []Config{
	config{1000, cRegisterBlueprints},
	config{1001, cSessionInit},
	config{1002, cRegisterTemplateRender},
}

func cRegisterBlueprints(a *App) error {
	for _, b := range a.ListBlueprints() {
		if !b.Registered() {
			b.Register()
		}
	}
	return nil
}

func cSessionInit(a *App) error {
	a.Init()
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
func Mode(mode string, value bool) Config {
	return DefaultConfig(func(a *App) error {
		return a.SetMode(mode, value)
	})
}

// Store returns a ConfigurationFn that adds key value items to the environment
// Store in the form of "key:value".
func Store(items ...string) Config {
	return DefaultConfig(func(a *App) error {
		for _, item := range items {
			v := strings.Split(item, ":")
			key, value := v[0], v[1]
			a.Add(key, value)
		}
		return nil
	})
}

// Extend returns a ConfigurationFn that adds the provided extension.Extensions
// to the app Environment.
func Extend(fxs ...extension.Extension) Config {
	return DefaultConfig(func(a *App) error {
		a.Extend(fxs...)
		return nil
	})
}

// Assets returns a ConfigurationFn adding the provided AssetFS to the app
// Environment Assets.
func Assets(as ...asset.AssetFS) Config {
	return DefaultConfig(func(a *App) error {
		a.SetAssetFS(as...)
		return nil
	})
}
