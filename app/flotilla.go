package app

import (
	"fmt"
	"net/http"

	"github.com/thrisp/flotilla/blueprint"
	"github.com/thrisp/flotilla/engine"
)

type App struct {
	name string
	engine.Engine
	Configuration
	Environment
	blueprint.Blueprints
}

// Empty returns an App instance with nothing but a name.
func Empty(name string) *App {
	return &App{name: name}
}

// Base returns an intialized App with provided Configuration immediately applied.
func Base(name string, conf ...ConfigurationFn) *App {
	app := Empty(name)
	runConfigure(app, conf...)
	runConfigure(app, configureImmediate...)
	return app
}

// New returns a Base initialized App with default blueprint and static directory
// plus any provided configuration.
func New(name string, conf ...ConfigurationFn) *App {
	app := Base(name)
	//app.STATIC("static")
	app.Configuration = defaultConfiguration(app, conf...)
	return app
}

// Returns the App name as a string.
func (a *App) Name() string {
	return a.name
}

func (a *App) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	a.Engine.ServeHTTP(rw, rq)
}

func (a *App) Run(addr string) {
	if !a.Configured() {
		if err := a.Configure(); err != nil {
			panic(fmt.Sprintf("[FLOTILLA] app could not be configured properly:\n%s", err))
		}
	}
	if err := http.ListenAndServe(addr, a); err != nil {
		panic(err)
	}
}
