package flotilla

import (
	"fmt"
	"net/http"
)

type (
	Manage func(*Ctx)

	App struct {
		name string
		*engine
		*Config
		*Env
		*Messaging
		*Blueprint
	}
)

// Empty returns an App instance with nothing but a name.
func Empty(name string) *App {
	return &App{name: name}
}

// Base returns an intialized App with no configuration.
func Base(name string) *App {
	app := Empty(name)
	app.engine = newEngine(app)
	app.Env = newEnv(app)
	app.Messaging = newMessaging()
	return app
}

// New returns a Base initialized App with default plus any provided configuration.
func New(name string, conf ...Configuration) *App {
	app := Base(name)
	app.Blueprint = NewBlueprint("/")
	app.STATIC("static")
	app.Config = newConfig(conf...)
	return app
}

func (a *App) Name() string {
	return a.name
}

func (a *App) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	a.engine.ServeHTTP(rw, req)
}

func (a *App) Run(addr string) {
	if !a.Configured {
		if err := a.Configure(a.Configuration...); err != nil {
			panic(fmt.Sprintf("[FLOTILLA] app could not be configured properly: %s", err))
		}
	}
	if err := http.ListenAndServe(addr, a); err != nil {
		panic(err)
	}
}
