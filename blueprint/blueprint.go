package blueprint

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/route"
	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/static"
	"github.com/thrisp/flotilla/status"
	//"github.com/thrisp/flotilla/xrr"
)

// Blueprint is an interface for common route bundling in a flotilla app.
type Blueprint interface {
	SetupState
	Handles
	Makes
	Prefix() string
	route.Routes
	Exists(*route.Route) bool
	New(string, ...state.Manage) Blueprint
	Use(...state.Manage)
	UseAt(int, ...state.Manage)
	Managers() []state.Manage
	Parent(...Blueprint)
	Descendents() []Blueprint
	MethodManager
	status.Statusr
}

type SetupState interface {
	Register()
	Registered() bool
	Held() []*route.Route
}

type setupstate struct {
	registered bool
	deferred   []func()
	held       []*route.Route
}

func (s *setupstate) Register() {
	s.runDeferred()
	s.registered = true
}

func (s *setupstate) runDeferred() {
	for _, fn := range s.deferred {
		fn()
	}
	s.deferred = nil
}

func (s *setupstate) Registered() bool {
	return s.registered
}

func (s *setupstate) Held() []*route.Route {
	return s.held
}

type HandleFn func(string, string, engine.Rule)

type Handles interface {
	Handling(string, string, engine.Rule)
}

type handles struct {
	handle HandleFn
}

func NewHandles(hf HandleFn) Handles {
	return &handles{hf}
}

func (h *handles) Handling(method string, path string, rule engine.Rule) {
	h.handle(method, path, rule)
}

type Makes interface {
	Making() state.Make
}

type makes struct {
	makes state.Make
}

func NewMakes(m state.Make) Makes {
	return &makes{m}
}

func (m *makes) Making() state.Make {
	return m.makes
}

type blueprint struct {
	*setupstate
	Handles
	Makes
	prefix      string
	descendents []Blueprint
	route.Routes
	managers []state.Manage
	MethodManager
	status.Statusr
}

// New returns a new Blueprint with the provided string prefix, Handles & Makes
// interfaces.
func New(prefix string, h Handles, m Makes) Blueprint {
	return newBlueprint(prefix, h, m)
}

func newBlueprint(prefix string, h Handles, m Makes) *blueprint {
	return &blueprint{
		setupstate: &setupstate{},
		Handles:    h,
		Makes:      m,
		prefix:     prefix,
		Routes:     route.NewRoutes(),
		Statusr:    status.New(m),
	}
}

func (b *blueprint) Prefix() string {
	return b.prefix
}

func (b *blueprint) Exists(rt *route.Route) bool {
	for _, r := range b.Routes.All() {
		if (rt.Path == r.Path) && (rt.Method == r.Method) {
			return true
		}
	}
	return false
}

func (b *blueprint) pathFor(path string) string {
	joined := filepath.ToSlash(filepath.Join(b.prefix, path))
	// Append a '/' if the last component had one, but only if it's not there already
	if len(path) > 0 && path[len(path)-1] == '/' && joined[len(joined)-1] != '/' {
		return joined + "/"
	}
	return joined
}

func combineManagers(b Blueprint, managers []state.Manage) []state.Manage {
	h := make([]state.Manage, 0)
	h = append(h, b.Managers()...)
	for _, manage := range managers {
		if !manageExists(h, manage) {
			h = append(h, manage)
		}
	}
	return h
}

// New returns a new Blueprint as a child of the parent Blueprint.
func (b *blueprint) New(component string, managers ...state.Manage) Blueprint {
	prefix := b.pathFor(component)
	newb := newBlueprint(prefix, b.Handles, b.Makes)
	newb.managers = combineManagers(b, managers)
	b.descendents = append(b.descendents, newb)
	return newb
}

func isFunc(fn interface{}) bool {
	return reflect.ValueOf(fn).Kind() == reflect.Func
}

func equalFunc(a, b interface{}) bool {
	if !isFunc(a) || !isFunc(b) {
		panic("flotilla : funcEqual -- type error!")
	}
	av := reflect.ValueOf(&a).Elem()
	bv := reflect.ValueOf(&b).Elem()
	return av.InterfaceData() == bv.InterfaceData()
}

func manageExists(inside []state.Manage, outside state.Manage) bool {
	for _, fn := range inside {
		if equalFunc(fn, outside) {
			return true
		}
	}
	return false
}

func (b *blueprint) Use(managers ...state.Manage) {
	for _, manage := range managers {
		if !manageExists(b.managers, manage) {
			b.managers = append(b.managers, manage)
		}
	}
}

func (b *blueprint) UseAt(index int, managers ...state.Manage) {
	if index > len(b.managers) {
		b.Use(managers...)
		return
	}

	var newh []state.Manage

	for _, manage := range managers {
		if !manageExists(b.managers, manage) {
			newh = append(newh, manage)
		}
	}

	before := b.managers[:index]
	after := append(newh, b.managers[index:]...)
	b.managers = append(before, after...)
}

func (b *blueprint) add(r *route.Route) {
	b.Routes.SetRoute(r)
}

func (b *blueprint) hold(r *route.Route) {
	b.held = append(b.held, r)
}

func (b *blueprint) push(register func(), rt *route.Route) {
	if b.registered {
		register()
	} else {
		if rt != nil {
			b.hold(rt)
		}
		b.deferred = append(b.deferred, register)
	}
}

// Parent will add the provided blueprints as descendents of the Blueprint.
func (b *blueprint) Parent(bs ...Blueprint) {
	b.descendents = append(b.descendents, bs...)
}

// Descendents returns an array of Blueprint as direct decendents of the calling Blueprint.
func (b *blueprint) Descendents() []Blueprint {
	return b.descendents
}

func (b *blueprint) Managers() []state.Manage {
	return b.managers
}

func reManage(rt *route.Route, b Blueprint) {
	var ms []state.Manage
	existing := rt.Managers
	rt.Managers = nil
	ms = append(ms, combineManagers(b, existing)...)
	rt.Managers = ms
}

//
func registerRouteConf(b *blueprint) route.RouteConf {
	return func(rt *route.Route) error {
		reManage(rt, b)
		rt.Path = b.pathFor(rt.Base)
		rt.Registered = true
		rt.Makes = b.Makes
		return nil
	}
}

//
type MethodManager interface {
	Manage(rt *route.Route)
	GET(string, ...state.Manage)
	POST(string, ...state.Manage)
	DELETE(string, ...state.Manage)
	PATCH(string, ...state.Manage)
	PUT(string, ...state.Manage)
	OPTIONS(string, ...state.Manage)
	HEAD(string, ...state.Manage)
	STATIC(static.Static, string, ...string)
	STATUS(int, ...state.Manage)
}

// Manage adds a route to the Blueprint.
func (b *blueprint) Manage(rt *route.Route) {
	register := func() {
		rt.Configure(registerRouteConf(b))
		if !b.Exists(rt) {
			rt.Managers = append([]state.Manage{b.statusExtension}, rt.Managers...)
			b.add(rt)
			b.Handling(rt.Method, rt.Path, rt.Rule)
		}
	}
	b.push(register, rt)
}

//
func (b *blueprint) GET(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("GET", path, managers)))
}

//
func (b *blueprint) POST(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("POST", path, managers)))
}

//
func (b *blueprint) DELETE(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("DELETE", path, managers)))
}

//
func (b *blueprint) PATCH(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("PATCH", path, managers)))
}

//
func (b *blueprint) PUT(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("PUT", path, managers)))
}

//
func (b *blueprint) OPTIONS(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("OPTIONS", path, managers)))
}

//
func (b *blueprint) HEAD(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("HEAD", path, managers)))
}

func dropTrailing(path string, trailing string) string {
	if fp := strings.Split(path, "/"); fp[len(fp)-1] == trailing {
		return strings.Join(fp[0:len(fp)-1], "/")
	}
	return path
}

//
func (b *blueprint) STATIC(s static.Static, path string, dirs ...string) {
	if len(dirs) > 0 {
		for _, dir := range dirs {
			b.push(func() { s.StaticDirs(dropTrailing(dir, "*filepath")) }, nil)
		}
	}
	register := func() {
		rt := route.New(route.StaticRouteConf("GET", path, []state.Manage{s.StaticManage}))
		rt.Configure(registerRouteConf(b))
		b.add(rt)
		b.Handling(rt.Method, rt.Path, rt.Rule)
	}
	b.push(register, nil)
}

func formatStatusPath(code, prefix string) string {
	if prefix == "/" {
		return fmt.Sprintf("/%s/*filepath", code)
	}
	return fmt.Sprintf("/%s/%s/*filepath", code, prefix)
}

func (b *blueprint) statusExtension(s state.State) {
	s.Extend(b.StateStatusExtension())
}

//
func (b *blueprint) STATUS(code int, managers ...state.Manage) {
	b.SetRawStatus(code, managers...)
	b.push(func() {
		b.Handling(
			"STATUS",
			formatStatusPath(strconv.Itoa(code), b.prefix),
			b.StatusRule(),
		)
	},
		nil)
}
