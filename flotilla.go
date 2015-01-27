package flotilla

import (
	"fmt"
	"net/http"
	"sync"
)

type (
	App struct {
		p    sync.Pool
		name string
		*engine
		*Config
		*Env
		*Messaging
		*Blueprint
	}
)

// Base returns an intialized App with no configuration.
func Base(name string) *App {
	app := &App{name: name,
		engine:    newEngine(),
		Env:       EmptyEnv(),
		Messaging: newMessaging(),
	}
	app.p.New = app.newCtx
	return app
}

// New returns a Base initialized App with default plus any provided configuration.
func New(name string, conf ...Configuration) *App {
	app := Base(name)
	app.BaseEnv()
	app.Blueprint = NewBlueprint("/")
	app.STATIC("static")
	app.Config = newConfig(conf...)
	return app
}

func (a *App) Name() string {
	return a.name
}

func (a *App) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx, cancel := a.getCtx(rw, req)
	a.engine.serve(ctx)
	cancel(a, ctx)
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
