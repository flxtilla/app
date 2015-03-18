package flotilla

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/xrr"
)

var (
	FlotillaPath     string
	workingPath      string
	workingStatic    string
	workingTemplates string
)

type (
	// Modes configure specific modes for later reference in the App; unless set,
	// an App defaults to Development true, Production false, and Testing false.
	Modes struct {
		Development bool
		Production  bool
		Testing     bool
	}

	// Env is the primary environment reference for an App.
	Env struct {
		Mode *Modes
		Store
		SessionManager *session.Manager
		Assets
		Staticor
		Templator
		fxtensions    map[string]Fxtension
		tplfunctions  map[string]interface{}
		ctxprocessors map[string]reflect.Value
		customstatus  map[int]*status
		mkctx         MakeCtxFunc
	}
)

func newEnv(a *App) *Env {
	e := &Env{Mode: defaultModes(), Store: defaultStore()}
	e.AddFxtensions(BuiltInExtensions(a)...)
	return e
}

func defaultModes() *Modes {
	return &Modes{true, false, false}
}

// SetMode sets the provided Modes witht the provided boolean value.
// e.g. env.SetMode("Production", true)
func (env *Env) SetMode(mode string, value bool) error {
	m := reflect.ValueOf(env.Mode).Elem().FieldByName(mode)
	if m.CanSet() {
		m.SetBool(value)
		return nil
	}
	return xrr.NewError("env could not be set to %s", mode)
}

// CurrentMode returns Modes specific to the App the provided Ctx is running within.
func CurrentMode(c Ctx) *Modes {
	m, _ := c.Call("mode")
	return m.(*Modes)
}

// AddCtxProcesor adds a ctxprocessor function with the name and function interface.
func (env *Env) AddCtxProcessor(name string, fn interface{}) {
	if env.ctxprocessors == nil {
		env.ctxprocessors = make(map[string]reflect.Value)
	}
	env.ctxprocessors[name] = valueFunc(fn)
}

// Add AddCtxProcessors adds ctxprocessor functions from a map of string keyed interfaces.
func (env *Env) AddCtxProcessors(fns map[string]interface{}) {
	for k, v := range fns {
		env.AddCtxProcessor(k, v)
	}
}

// AddExtensions adds extension functions from a map of string keyed interfaces.
func (env *Env) AddFxtensions(fxs ...Fxtension) error {
	var err error
	if env.fxtensions == nil {
		env.fxtensions = make(map[string]Fxtension)
	}
	for _, fx := range fxs {
		err = validFxtension(fx)
		if err != nil {
			return err
		}
		env.fxtensions[fx.Name()] = fx
	}
	return err
}

// AddTplFunc adds a template function with the name and function interface.
func (env *Env) AddTplFunc(name string, fn interface{}) {
	if env.tplfunctions == nil {
		env.tplfunctions = make(map[string]interface{})
	}
	env.tplfunctions[name] = fn
}

// AddTplFuncs adds template functions from a map of string keyed interfaces.
func (env *Env) AddTplFuncs(fns map[string]interface{}) {
	for k, v := range fns {
		env.AddTplFunc(k, v)
	}
}

func (env *Env) defaultsessionconfig() string {
	secret := env.Store["SECRET_KEY"].Value
	cookie_name := env.Store["SESSION_COOKIENAME"].Value
	session_lifetime := env.Store["SESSION_LIFETIME"].Int64()
	prvdrcfg := fmt.Sprintf(`"ProviderConfig":"{\"maxage\": %d,\"cookieName\":\"%s\",\"securityKey\":\"%s\"}"`, session_lifetime, cookie_name, secret)
	return fmt.Sprintf(`{"cookieName":"%s","enableSetCookie":false,"gclifetime":3600, %s}`, cookie_name, prvdrcfg)
}

func (env *Env) defaultsessionmanager() *session.Manager {
	d, err := session.NewManager("cookie", env.defaultsessionconfig())
	if err != nil {
		panic(fmt.Sprintf("Problem with [FLOTILLA] default session manager: %s", err))
	}
	return d
}

// SessionInit intializes the SessionManager stored with the Env.
func (env *Env) SessionInit() {
	if env.SessionManager == nil {
		env.SessionManager = env.defaultsessionmanager()
	}
	go env.SessionManager.GC()
}

// CustomStatus sets a custom status keyed by integer within the Env reference.
func (env *Env) CustomStatus(s *status) {
	if env.customstatus == nil {
		env.customstatus = make(map[int]*status)
	}
	env.customstatus[s.code] = s
}

func init() {
	workingPath, _ = os.Getwd()
	workingPath, _ = filepath.Abs(workingPath)
	workingStatic, _ = filepath.Abs("./static")
	workingTemplates, _ = filepath.Abs("./templates")
	FlotillaPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
}
