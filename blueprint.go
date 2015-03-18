package flotilla

import (
	"path/filepath"
	"strconv"

	"github.com/thrisp/flotilla/xrr"
)

type (
	setupstate struct {
		registered bool
		deferred   []func()
		held       []*Route
	}

	// Blueprint is a common grouping of Routes.
	Blueprint struct {
		*setupstate
		app      *App
		children []*Blueprint
		Prefix   string
		Routes
		Managers []Manage
		MakeCtx  MakeCtxFunc
	}
)

// Blueprints returns a flat array of Blueprints attached to the App.
func (a *App) Blueprints() []*Blueprint {
	type IterC func(bs []*Blueprint, fn IterC)

	var bps []*Blueprint

	bps = append(bps, a.Blueprint)

	iter := func(bs []*Blueprint, fn IterC) {
		for _, x := range bs {
			bps = append(bps, x)
			fn(x.children, fn)
		}
	}

	iter(a.children, iter)

	return bps
}

// Given any number of Blueprints, RegisterBlueprints registers each with the App.
func (a *App) RegisterBlueprints(blueprints ...*Blueprint) {
	for _, blueprint := range blueprints {
		blueprint.Register(a)
		if _, exists := a.ExistingBlueprint(blueprint.Prefix); !exists {
			a.children = append(a.children, blueprint)
		}
	}
}

func (a *App) ExistingBlueprint(prefix string) (*Blueprint, bool) {
	for _, b := range a.Blueprints() {
		if b.Prefix == prefix {
			return b, true
		}
	}
	return nil, false
}

// Mount attaches each provided Blueprint to the given string mount point, optionally
// inheriting from and setting the app primary Blueprint as parent to the given Blueprints.
func (a *App) Mount(point string, blueprints ...*Blueprint) error {
	var b []*Blueprint
	for _, blueprint := range blueprints {
		if blueprint.registered {
			return xrr.NewError("only unregistered blueprints may be mounted; %s is already registered", blueprint.Prefix)
		}

		newprefix := filepath.ToSlash(filepath.Join(point, blueprint.Prefix))

		nbp := a.NewBlueprint(newprefix)

		for _, rt := range blueprint.setupstate.held {
			nbp.Manage(rt)
		}

		b = append(b, nbp)
	}
	a.RegisterBlueprints(b...)
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
	return &Blueprint{
		setupstate: &setupstate{},
		Prefix:     prefix,
		Routes:     make(Routes),
	}
}

// NewBlueprint returns a new Blueprint as a child of the parent Blueprint.
func (b *Blueprint) NewBlueprint(component string, managers ...Manage) *Blueprint {
	prefix := b.pathFor(component)

	newb := NewBlueprint(prefix)
	newb.Managers = b.combineManagers(managers)

	b.children = append(b.children, newb)

	return newb
}

// Register intergrates App information with the Blueprint.
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

func (b *Blueprint) combineManagers(managers []Manage) []Manage {
	s := len(b.Managers) + len(managers)
	h := make([]Manage, 0, s)
	h = append(h, b.Managers...)
	h = append(h, managers...)
	return h
}

func (b *Blueprint) manageExists(outside Manage) bool {
	for _, inside := range b.Managers {
		if equalFunc(inside, outside) {
			return true
		}
	}
	return false
}

// Use adds Manage functions to the Blueprint.
func (b *Blueprint) Use(managers ...Manage) {
	for _, manage := range managers {
		if !b.manageExists(manage) {
			b.Managers = append(b.Managers, manage)
		}
	}
}

// UseAt adds Manage functions to the Blueprint at the provided index.
func (b *Blueprint) UseAt(index int, managers ...Manage) {
	if index > len(b.Managers) {
		b.Use(managers...)
		return
	}

	var newh []Manage

	for _, manage := range managers {
		if !b.manageExists(manage) {
			newh = append(newh, manage)
		}
	}

	before := b.Managers[:index]
	after := append(newh, b.Managers[index:]...)
	b.Managers = append(before, after...)
}

func (b *Blueprint) routeExists(route *Route) bool {
	for _, r := range b.Routes {
		if route.path == r.path {
			return true
		}
	}
	return false
}

func (b *Blueprint) add(r *Route) {
	if r.Name != "" {
		b.Routes[r.Name] = r
	} else {
		b.Routes[r.Named()] = r
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

func (b *Blueprint) mkctxfunc() MakeCtxFunc {
	if b.MakeCtx == nil {
		return b.app.Ctx()
	}
	return b.MakeCtx
}

func (b *Blueprint) register(route *Route) {
	route.blueprint = b
	route.managers = b.combineManagers(route.managers)
	route.path = b.pathFor(route.base)
	route.registered = true
	route.MakeCtx = b.mkctxfunc()
}

// Manage adds a route to the Blueprint.
func (b *Blueprint) Manage(route *Route) {
	register := func() {
		b.register(route)
		if !b.routeExists(route) {
			b.add(route)
			b.app.Handle(route.method, route.path, route.rule)
		}
	}
	b.push(register, route)
}

func (b *Blueprint) GET(path string, managers ...Manage) {
	b.Manage(NewRoute("GET", path, false, managers))
}

func (b *Blueprint) POST(path string, managers ...Manage) {
	b.Manage(NewRoute("POST", path, false, managers))
}

func (b *Blueprint) DELETE(path string, managers ...Manage) {
	b.Manage(NewRoute("DELETE", path, false, managers))
}

func (b *Blueprint) PATCH(path string, managers ...Manage) {
	b.Manage(NewRoute("PATCH", path, false, managers))
}

func (b *Blueprint) PUT(path string, managers ...Manage) {
	b.Manage(NewRoute("PUT", path, false, managers))
}

func (b *Blueprint) OPTIONS(path string, managers ...Manage) {
	b.Manage(NewRoute("OPTIONS", path, false, managers))
}

func (b *Blueprint) HEAD(path string, managers ...Manage) {
	b.Manage(NewRoute("HEAD", path, false, managers))
}

func (b *Blueprint) STATIC(path string) {
	b.push(func() { b.app.StaticDirs(dropTrailing(path, "*filepath")) }, nil)
	register := func() {
		route := NewRoute("GET", path, true, []Manage{b.app.Staticor.Manage})
		b.register(route)
		b.add(route)
		b.app.Handle(route.method, route.path, route.rule)
	}
	b.push(register, nil)
}

func (b *Blueprint) STATUS(code int, managers ...Manage) {
	b.push(func() { b.app.Handle("STATUS", strconv.Itoa(code), CustomStatusRule(b.app, code, managers...)) }, nil)
}
