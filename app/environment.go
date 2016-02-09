package app

import (
	"os"
	"path/filepath"

	"github.com/thrisp/flotilla/asset"
	"github.com/thrisp/flotilla/extension"
	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/static"
	"github.com/thrisp/flotilla/store"
	"github.com/thrisp/flotilla/template"
)

// Environemnt is an interface for central storage and access of crucial app
// functionality. These include State creation, Logging, and app Mode
// determination and setting, in addition to package external Assets,
// Extension, Session, Static, Store, and Templates.
type Environment interface {
	Statr
	Logr
	Modr
	asset.Assets
	extension.Extension
	session.Sessions
	static.Static
	store.Store
	template.Templates
}

type environment struct {
	Logr
	Modr
	Statr
	asset.Assets
	extension.Extension
	session.Sessions
	static.Static
	store.Store
	template.Templates
}

func newEnvironment(app *App) Environment {
	st := defaultStore()
	as := asset.New()
	return &environment{
		Logr:      DefaultLogr(),
		Modr:      DefaultModr(),
		Statr:     DefaultStatr(),
		Assets:    as,
		Extension: BuiltInExtension(app),
		Store:     st,
		Sessions:  session.NewSessions(st),
		Static:    static.New(st, as),
		Templates: template.New(template.DefaultTemplatr(st, as)),
	}
}

func defaultStore() store.Store {
	s := store.New()
	s.Add("upload_size", "10000000")
	s.Add("secret_key", "Flotilla;Secret;Key;1")
	s.Add("session_cookiename", "session")
	s.Add("session_lifetime", "2629743")
	s.Add("working_path", workingPath)
	s.Add("flotilla_path", FlotillaPath)
	s.Add("static_directories", workingStatic)
	s.Add("template_directories", workingTemplates)
	return s
}

var (
	FlotillaPath     string
	workingPath      string
	workingStatic    string
	workingTemplates string
)

func init() {
	workingPath, _ = os.Getwd()
	workingPath, _ = filepath.Abs(workingPath)
	workingStatic, _ = filepath.Abs("./static")
	workingTemplates, _ = filepath.Abs("./templates")
	FlotillaPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
}
