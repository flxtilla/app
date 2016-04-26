package app

import (
	"fmt"
	"net/http"

	"github.com/flxtilla/cxre/blueprint"
	"github.com/flxtilla/cxre/engine"
)

// App is the cxre structure for a flotilla application, implementing the
// Engine, Configuration, Environment, and Blueprints interfaces.
type App struct {
	name string
	engine.Engine
	Configuration
	Environment
	blueprint.Blueprints
}

// Empty returns an App instance with the provided name.
func Empty(name string) *App {
	return &App{name: name}
}

// Base returns an intialized App with crucial and the provided
// ConfigurationFns immediately applied.
func Base(name string, conf ...Config) *App {
	app := Empty(name)
	configure(app, preConfig...)
	configure(app, conf...)
	return app
}

// New returns a default Base initialized App with static directory plus any
// provided configuration. Note: ConfigurationFn supplied to New are not run
// until the App is configured. If you need to modify certain base
// functionality(e.g.  custom engine, environment, or blueprints), start with
// Base instead of New.
func New(name string, conf ...Config) *App {
	app := Base(name)
	app.STATIC(app.Environment, "/static/*filepath", "static")
	app.Configuration = newConfiguration(app, conf...)
	return app
}

// Returns the App name as a string.
func (a *App) Name() string {
	return a.name
}

// ServeHTTP function for the App
func (a *App) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	a.Engine.ServeHTTP(rw, rq)
}

// Run checks the App is configured, configuring and panicing on errors, then
// starts the App listening at the provided address.
func (a *App) Run(addr string) {
	if !a.Configured() {
		if err := a.Configure(); err != nil {
			a.Panic(fmt.Sprintf("[FLOTILLA] app could not be configured properly:\n%s", err))
		}
	}
	if err := http.ListenAndServe(addr, a); err != nil {
		a.Panic(err)
	}
}
