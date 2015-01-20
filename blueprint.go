package flotilla

import "path/filepath"

type (
	setupstate struct {
		registered bool
		deferred   []func()
		held       []*Route
	}

	// A Blueprint gathers any number routes around a prefix and an array of
	// group specific handlers.
	Blueprint struct {
		*setupstate
		app      *App
		children []*Blueprint
		routes   Routes
		Prefix   string
		Handlers []Manage
	}
)

// Blueprints provides a flat array of Blueprint instances attached to the App.
func (app *App) Blueprints() []*Blueprint {
	type IterC func(bs []*Blueprint, fn IterC)

	var bps []*Blueprint

	bps = append(bps, app.Blueprint)

	iter := func(bs []*Blueprint, fn IterC) {
		for _, x := range bs {
			bps = append(bps, x)
			fn(x.children, fn)
		}
	}

	iter(app.children, iter)

	return bps
}

// RegisterBlueprints integrates the given blueprints with the App.
func (app *App) RegisterBlueprints(blueprints ...*Blueprint) {
	for _, blueprint := range blueprints {
		if existing, ok := app.existingBlueprint(blueprint.Prefix); ok {
			existing.Use(blueprint.Handlers...)
			app.MergeRoutes(existing, blueprint.routes)
		} else {
			app.children = append(app.children, blueprint)
			blueprint.Register(app)
		}
	}
}

func (app *App) existingBlueprint(prefix string) (*Blueprint, bool) {
	for _, b := range app.Blueprints() {
		if b.Prefix == prefix {
			return b, true
		}
	}
	return nil, false
}

// Mount takes an unregistered blueprint, registering and mounting the routes to
// the provided string mount point with a copy of the blueprint. If inherit is
// true, the blueprint becomes a child blueprint of app.Blueprint.
func (app *App) Mount(mount string, inherit bool, blueprints ...*Blueprint) error {
	var mbp *Blueprint
	var mbs []*Blueprint
	for _, blueprint := range blueprints {
		if blueprint.registered {
			return newError("only unregistered blueprints may be mounted; %s is already registered", blueprint.Prefix)
		}

		newprefix := filepath.ToSlash(filepath.Join(mount, blueprint.Prefix))

		if inherit {
			mbp = app.NewBlueprint(newprefix)
		} else {
			mbp = NewBlueprint(newprefix)
		}

		for _, route := range blueprint.held {
			mbp.Handle(route)
		}

		mbs = append(mbs, mbp)
	}
	app.RegisterBlueprints(mbs...)
	return nil
}

func (b *Blueprint) pathFor(path string) string {
	joined := filepath.ToSlash(filepath.Join(b.Prefix, path))
	// Append a '/' if the last component had one, but only if it's not there already
	if len(path) > 0 && path[len(path)-1] == '/' && joined[len(joined)-1] != '/' {
		return joined + "/"
	}
	return joined
}

// NewBlueprint returns a new Blueprint with the provided string prefix.
func NewBlueprint(prefix string) *Blueprint {
	return &Blueprint{setupstate: &setupstate{},
		Prefix: prefix,
		routes: make(Routes),
	}
}

// New creates a new child Blueprint from the existing Blueprint.
func (b *Blueprint) NewBlueprint(component string, handlers ...Manage) *Blueprint {
	prefix := b.pathFor(component)

	newb := NewBlueprint(prefix)
	newb.Handlers = b.combineHandlers(handlers)

	b.children = append(b.children, newb)

	return newb
}

// Register will provide the app instance to the blueprint to finalize all deferred actions.
func (b *Blueprint) Register(a *App) {
	b.app = a
	b.runDeferred()
	b.registered = true
}

func (b *Blueprint) runDeferred() {
	for _, fn := range b.deferred {
		fn()
	}
	b.deferred = nil
}

func (b *Blueprint) combineHandlers(handlers []Manage) []Manage {
	s := len(b.Handlers) + len(handlers)
	h := make([]Manage, 0, s)
	h = append(h, b.Handlers...)
	h = append(h, handlers...)
	return h
}

func (b *Blueprint) handlerExists(outside Manage) bool {
	for _, inside := range b.Handlers {
		if equalFunc(inside, outside) {
			return true
		}
	}
	return false
}

// Use adds any number of Manage to the Blueprint which will be run before
// route handlers for all Route attached to the Blueprint.
func (b *Blueprint) Use(handlers ...Manage) {
	for _, handler := range handlers {
		if !b.handlerExists(handler) {
			b.Handlers = append(b.Handlers, handler)
		}
	}
}

// UseAt adds any number of Manage to the Blueprint at the given index, for
// when you must control the position in relation to other middleware.
func (b *Blueprint) UseAt(index int, handlers ...Manage) {
	if index > len(b.Handlers) {
		b.Use(handlers...)
		return
	}

	var newh []Manage

	for _, handler := range handlers {
		if !b.handlerExists(handler) {
			newh = append(newh, handler)
		}
	}

	before := b.Handlers[:index]
	after := append(newh, b.Handlers[index:]...)
	b.Handlers = append(before, after...)
}

func (b *Blueprint) add(r *Route) {
	if r.Name != "" {
		b.routes[r.Name] = r
	} else {
		b.routes[r.Named()] = r
	}
}

func (b *Blueprint) hold(r *Route) {
	b.held = append(b.held, r)
}

func (b *Blueprint) push(register func(), route *Route) {
	if b.registered {
		register()
	} else {
		if route != nil {
			b.hold(route)
		}
		b.deferred = append(b.deferred, register)
	}
}

func (b *Blueprint) register(route *Route) {
	route.blueprint = b
	route.handlers = b.combineHandlers(route.handlers)
	route.path = b.pathFor(route.base)
	route.registered = true
}

// Handle registers new handlers and/or existing handlers with a constructed Route.
// For GET, POST, DELETE, PATCH, PUT, OPTIONS, and HEAD requests the respective
// shortcut functions can be used by specifying path & handlers.
func (b *Blueprint) Handle(route *Route) {
	register := func() {
		b.register(route)
		b.add(route)
		b.app.manage(route.method, route.path, route.handle)
	}
	b.push(register, route)
}

func (b *Blueprint) GET(path string, handlers ...Manage) {
	b.Handle(NewRoute("GET", path, false, handlers))
}

func (b *Blueprint) POST(path string, handlers ...Manage) {
	b.Handle(NewRoute("POST", path, false, handlers))
}

func (b *Blueprint) DELETE(path string, handlers ...Manage) {
	b.Handle(NewRoute("DELETE", path, false, handlers))
}

func (b *Blueprint) PATCH(path string, handlers ...Manage) {
	b.Handle(NewRoute("PATCH", path, false, handlers))
}

func (b *Blueprint) PUT(path string, handlers ...Manage) {
	b.Handle(NewRoute("PUT", path, false, handlers))
}

func (b *Blueprint) OPTIONS(path string, handlers ...Manage) {
	b.Handle(NewRoute("OPTIONS", path, false, handlers))
}

func (b *Blueprint) HEAD(path string, handlers ...Manage) {
	b.Handle(NewRoute("HEAD", path, false, handlers))
}

// STATIC adds a Static route handled by the app, based on the blueprint prefix.
func (b *Blueprint) STATIC(path string) {
	b.push(func() { b.app.StaticDirs(dropTrailing(path, "*filepath")) }, nil)
	b.Handle(NewRoute("GET", path, true, []Manage{handleStatic}))
}

/*
// Custom HttpStatus for the group, set and called from engine HttpStatuses
func (b *Blueprint) StatusHandle(code int, handlers ...Manage) {
	statushandler := func(c context.Context) {
		curr := c.Value("Current").(Current)
		statusCtx := b.app.tmpCtx(curr.Writer(), curr.Request())
		s := len(handlers)
		for i := 0; i < s; i++ {
			handlers[i](statusCtx)
		}
		for _, fn := range statusCtx.deferred {
			fn(statusCtx)
		}
	}
	register := func() {
		b.app.TakeStatus(code, statushandler)
	}
	b.push(register, nil)
}
*/
