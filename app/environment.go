package app

import (
	"os"
	"path/filepath"

	"github.com/thrisp/flotilla/assets"
	"github.com/thrisp/flotilla/extension"
	"github.com/thrisp/flotilla/log"
	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/static"
	"github.com/thrisp/flotilla/status"
	"github.com/thrisp/flotilla/store"
	"github.com/thrisp/flotilla/template"
)

type Environment interface {
	Modes
	store.Store
	assets.Assets
	static.Static
	template.Templates
	extension.Fxtension
	session.Sessions
	status.Statuses
	log.Logger
}

type environment struct {
	Modes
	store.Store
	assets.Assets
	static.Static
	template.Templates
	extension.Fxtension
	session.Sessions
	status.Statuses
	log.Logger
}

func newEnvironment(app *App) Environment {
	st := defaultStore()
	as := assets.New()
	return &environment{
		Modes:     defaultModes(),
		Store:     st,
		Assets:    as,
		Static:    static.New(st, as),
		Templates: template.New(template.DefaultTemplatr(st, as)),
		Fxtension: BuiltInExtension(app),
		Sessions:  session.NewSessions(st),
		Statuses:  status.New(),
		Logger:    log.New(os.Stdout, log.LInfo, log.DefaultTextFormatter()),
	}
}

func defaultStore() store.Store {
	s := store.New()
	s.Add("upload_size", "10000000")
	s.Add("secret_key", "Flotilla;Secret;Key;1")
	s.Add("session_cookiename", "session")
	s.Add("session_lifetime", "2629743")
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
