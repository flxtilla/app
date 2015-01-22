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
		*Blueprint
		*Messaging
	}
)

func Empty(name string) *App {
	app := &App{name: name,
		engine:    defaultEngine(),
		Env:       EmptyEnv(),
		Messaging: newMessaging(),
	}
	app.p.New = app.newCtx
	return app
}

func New(name string, conf ...Configuration) *App {
	app := Empty(name)
	app.BaseEnv()
	app.Config = defaultConfig()
	app.Blueprint = NewBlueprint("/")
	app.STATIC("static")
	app.Configured = false
	app.Configuration = conf
	return app
}

func (a *App) Name() string {
	return a.name
}

func (a *App) rcvr(c *Ctx) {
	if rcv := recover(); rcv != nil {
		p := newError(fmt.Sprintf("%s", rcv))
		c.errorTyped(p, ErrorTypePanic, stack(3))
		c.Status(500)
		a.putCtx(c)
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rslt := a.lookup(req.Method, req.URL.Path)
	ctx := a.getCtx(w, req, rslt)
	defer a.rcvr(ctx)
	rslt.handler(ctx)
	a.putCtx(ctx)
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
