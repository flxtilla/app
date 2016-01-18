package blueprint

import (
	"path/filepath"
	"reflect"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/route"
	"github.com/thrisp/flotilla/state"
	//"github.com/thrisp/flotilla/xrr"
)

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
}

// New returns a new Blueprint with the provided string prefix and HandleFn.
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
	s := len(b.Managers()) + len(managers)
	h := make([]state.Manage, 0, s)
	h = append(h, b.Managers()...)
	h = append(h, managers...)
	return h
}

// *blueprint.New returns a new Blueprint as a child of the parent Blueprint.
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

func (b *blueprint) manageExists(outside state.Manage) bool {
	for _, inside := range b.managers {
		if equalFunc(inside, outside) {
			return true
		}
	}
	return false
}

func (b *blueprint) Use(managers ...state.Manage) {
	for _, manage := range managers {
		if !b.manageExists(manage) {
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
		if !b.manageExists(manage) {
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

func (b *blueprint) Parent(bs ...Blueprint) {
	b.descendents = append(b.descendents, bs...)
}

func (b *blueprint) Descendents() []Blueprint {
	return b.descendents
}

func (b *blueprint) Managers() []state.Manage {
	return b.managers
}

func registerRouteConf(b *blueprint) route.RouteConf {
	return func(rt *route.Route) error {
		rt.Managers = combineManagers(b, rt.Managers)
		rt.Path = b.pathFor(rt.Base)
		rt.Registered = true
		rt.Makes = b.Makes
		return nil
	}
}

type MethodManager interface {
	Manage(rt *route.Route)
	GET(string, ...state.Manage)
	POST(string, ...state.Manage)
	DELETE(string, ...state.Manage)
	PATCH(string, ...state.Manage)
	PUT(string, ...state.Manage)
	OPTIONS(string, ...state.Manage)
	HEAD(string, ...state.Manage)
	//STATIC(string)
	//STATUS(int, ...state.Manage)
}

// Manage adds a route to the Blueprint.
func (b *blueprint) Manage(rt *route.Route) {
	register := func() {
		rt.Configure(registerRouteConf(b))
		if !b.Exists(rt) {
			b.add(rt)
			b.Handling(rt.Method, rt.Path, rt.Rule)
		}
	}
	b.push(register, rt)
}

func (b *blueprint) GET(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("GET", path, managers)))
}

func (b *blueprint) POST(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("POST", path, managers)))
}

func (b *blueprint) DELETE(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("DELETE", path, managers)))
}

func (b *blueprint) PATCH(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("PATCH", path, managers)))
}

func (b *blueprint) PUT(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("PUT", path, managers)))
}

func (b *blueprint) OPTIONS(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("OPTIONS", path, managers)))
}

func (b *blueprint) HEAD(path string, managers ...state.Manage) {
	b.Manage(route.New(route.DefaultRouteConf("HEAD", path, managers)))
}

//func (b *blueprint) STATIC(path string) {
//b.push(func() { /*b.app.StaticDirs(dropTrailing(path, "*filepath"))*/ }, nil)
//register := func() {
//	rt := route.New(route.StaticRouteConf("GET", path, []state.Manage{b.StaticManage()}))
//	rt.Configure(registerRouteConf(b))
//	b.add(rt)
//	b.HandleFn()(rt.Method, rt.Path, rt.Rule)
//}
//b.push(register, nil)
//}

//func (b *blueprint) STATUS(code int, managers ...state.Manage) {
//b.push(func() {
//b.HandleFn()(
//	"STATUS",
//	strconv.Itoa(code),
//	b.StatusRule(),
//) //(b.app, code, managers...))
//},
//	nil)
//}
