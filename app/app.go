package app

import (
	"fmt"
	"net/http"

	"github.com/thrisp/flotilla/blueprint"
	"github.com/thrisp/flotilla/engine"
)

// App is the core structure for a flotilla application, implementing the
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

// Base returns an intialized App with the provided ConfigurationFns
// immediately applied.
func Base(name string, conf ...ConfigurationFn) *App {
	app := Empty(name)
	runConfigure(app, conf...)
	runConfigure(app, configureImmediate...)
	return app
}

// New returns a Base initialized App with static directory plus any provided
// configuration. Note: ConfigurationFn supplied to New are not run until the
// App is configured.
func New(name string, conf ...ConfigurationFn) *App {
	app := Base(name)
	app.STATIC(app.Environment, "/static/*filepath", "static")
	app.Configuration = defaultConfiguration(app, conf...)
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
