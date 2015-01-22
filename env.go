package flotilla

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/thrisp/flotilla/session"
)

var (
	FlotillaPath     string
	workingPath      string
	workingStatic    string
	workingTemplates string
)

type (
	Modes struct {
		Development bool
		Production  bool
		Testing     bool
	}

	Env struct {
		Mode *Modes
		Store
		SessionManager *session.Manager
		Assets
		Staticor
		Templator
		extensions    map[string]interface{}
		tplfunctions  map[string]interface{}
		ctxprocessors map[string]interface{}
	}
)

func (e *Env) defaults() {
	e.Store.addDefault("upload", "size", "10000000")             // bytes
	e.Store.addDefault("secret", "key", "Flotilla;Secret;Key;1") // weak default value
	e.Store.addDefault("session", "cookiename", "session")
	e.Store.addDefault("session", "lifetime", "2629743")
	e.Store.add("static", "directories", workingStatic)
	e.Store.add("template", "directories", workingTemplates)
	e.Store.add("httpstatus", "html", "true")
}

func EmptyEnv() *Env {
	return &Env{Mode: &Modes{true, false, false},
		Store:         make(Store),
		extensions:    make(map[string]interface{}),
		tplfunctions:  make(map[string]interface{}),
		ctxprocessors: make(map[string]interface{}),
	}
}

func (env *Env) BaseEnv() {
	env.AddExtensions(builtinextensions)
	env.defaults()
}

func (env *Env) MergeEnv(other *Env) {
	env.MergeStore(other.Store)
	for _, fs := range other.Assets {
		env.Assets = append(env.Assets, fs)
	}
	env.StaticDirs(other.Store["STATIC_DIRECTORIES"].List()...)
	env.TemplateDirs(other.Store["TEMPLATE_DIRECTORIES"].List()...)
	env.AddExtensions(other.extensions)
}

func (env *Env) MergeStore(other Store) {
	for k, v := range other {
		if !v.defaultvalue {
			if _, ok := env.Store[k]; !ok {
				env.Store[k] = v
			}
		}
	}
}

func (env *Env) SetMode(mode string, value bool) error {
	m := reflect.ValueOf(env.Mode).Elem().FieldByName(mode)
	if m.CanSet() {
		m.SetBool(value)
		return nil
	}
	return newError("env could not be set to %s", mode)
}

func (env *Env) CtxProcessor(name string, fn interface{}) {
	env.ctxprocessors[name] = fn
}

func (env *Env) CtxProcessors(fns map[string]interface{}) {
	for k, v := range fns {
		env.CtxProcessor(k, v)
	}
}

func (env *Env) AddExtension(name string, fn interface{}) error {
	err := validExtension(fn)
	if err == nil {
		env.extensions[name] = fn
		return nil
	}
	return err
}

func (env *Env) AddExtensions(fns map[string]interface{}) error {
	for k, v := range fns {
		err := env.AddExtension(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (env *Env) AddTplFunc(name string, fn interface{}) {
	env.tplfunctions[name] = fn
}

func (env *Env) AddTplFuncs(fns map[string]interface{}) {
	for k, v := range fns {
		env.AddTplFunc(k, v)
	}
}

func (env *Env) defaultsessionconfig() string {
	secret := env.Store["SECRET_KEY"].Value
	cookie_name := env.Store["SESSION_COOKIENAME"].Value
	session_lifetime, _ := env.Store["SESSION_LIFETIME"].Int64()
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

func (env *Env) SessionInit() {
	if env.SessionManager == nil {
		env.SessionManager = env.defaultsessionmanager()
	}
	go env.SessionManager.GC()
}

func init() {
	workingPath, _ = os.Getwd()
	workingPath, _ = filepath.Abs(workingPath)
	workingStatic, _ = filepath.Abs("./static")
	workingTemplates, _ = filepath.Abs("./templates")
	FlotillaPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
}
