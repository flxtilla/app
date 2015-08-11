package flotilla

import (
	"mime/multipart"
	"net/http"
	"reflect"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/xrr"
)

// The Extensor interface provides access to varying functions in a Ctx.
type Extensor interface {
	Call(string, ...interface{}) (interface{}, error)
}

type extensor struct {
	ext map[string]reflect.Value
	ctx *ctx
}

func newextensor(fx map[string]Fxtension, c *ctx) Extensor {
	e := make(map[string]reflect.Value)
	for _, x := range fx {
		x.Set(e)
	}
	return &extensor{ext: e, ctx: c}
}

func (e *extensor) Call(name string, args ...interface{}) (interface{}, error) {
	var ctxargs []interface{}
	ctxargs = append(ctxargs, e.ctx)
	ctxargs = append(ctxargs, args...)
	return call(e.ext[name], ctxargs...)
}

type Fxtension interface {
	Name() string
	Set(map[string]reflect.Value)
}

type fxtension struct {
	name string
	fns  map[string]interface{}
}

var InvalidExtension = xrr.NewXrror("%q is not a valid Fxtension.").Out

func validFxtension(fx interface{}) error {
	if _, valid := fx.(Fxtension); !valid {
		InvalidExtension(fx)
	}
	return nil
}

func MakeFxtension(name string, fns map[string]interface{}) *fxtension {
	return &fxtension{
		name: name,
		fns:  fns,
	}
}

func (fx *fxtension) Name() string {
	return fx.name
}

func (fx *fxtension) Set(rv map[string]reflect.Value) {
	for k, v := range fx.fns {
		rv[k] = valueFunc(v)
	}
}

var responsefxtension = map[string]interface{}{
	"abort":           abort,
	"headernow":       headernow,
	"headerwrite":     headerwrite,
	"headermodify":    headermodify,
	"iswritten":       iswritten,
	"redirect":        redirect,
	"servefile":       servefile,
	"serveplain":      serveplain,
	"writetoresponse": writetoresponse,
}

var ResponseFxtension Fxtension = MakeFxtension("responsefxtension", responsefxtension)

func abort(c *ctx, code int) error {
	if code >= 0 {
		c.RW.WriteHeader(code)
		c.RW.WriteHeaderNow()
	}
	return nil
}

func headernow(c *ctx) error {
	c.RW.WriteHeaderNow()
	return nil
}

func headerwrite(c *ctx, code int, values ...[]string) error {
	if code >= 0 {
		c.RW.WriteHeader(code)
	}
	headermodify(c, "set", values...)
	return nil
}

func headermodify(c *ctx, action string, values ...[]string) error {
	switch action {
	case "set":
		for _, v := range values {
			c.RW.Header().Set(v[0], v[1])
		}
	default:
		for _, v := range values {
			c.RW.Header().Add(v[0], v[1])
		}
	}
	return nil
}

func iswritten(c *ctx) bool {
	return c.RW.Written()
}

var InvalidStatusCode = xrr.NewXrror("Cannot send a redirect with status code %d").Out

func redirect(c *ctx, code int, location string) error {
	if code >= 300 && code <= 308 {
		c.bounce(func(pc Ctx) {
			releasesession(c)
			http.Redirect(c.RW, c.Request, location, code)
			c.RW.WriteHeaderNow()
		})
		return nil
	} else {
		return InvalidStatusCode(code)
	}
}

func serveplain(c *ctx, code int, data string) error {
	c.push(func(pc Ctx) {
		headerwrite(c, code, []string{"Content-Type", "text/plain"})
		c.RW.Write([]byte(data))
	})
	return nil
}

func servefile(c *ctx, f http.File) error {
	fi, err := f.Stat()
	if err == nil {
		http.ServeContent(c.RW, c.Request, fi.Name(), fi.ModTime(), f)
	}
	return err
}

func writetoresponse(c *ctx, data string) error {
	c.RW.Write([]byte(data))
	return nil
}

var sessionfxtension = map[string]interface{}{
	"deletesession": deletesession,
	"getsession":    getsession,
	"release":       releasesession,
	"session":       returnsession,
	"setsession":    setsession,
	"start":         startsession,
}

var SessionFxtension Fxtension = MakeFxtension("sessionfxtension", sessionfxtension)

func deletesession(c *ctx, key string) error {
	return c.Session.Delete(key)
}

func getsession(c *ctx, key string) interface{} {
	return c.Session.Get(key)
}

func Session(c Ctx) session.SessionStore {
	s, _ := c.Call("session")
	return s.(session.SessionStore)
}

func returnsession(c *ctx) session.SessionStore {
	return c.Session
}

func releasesession(c *ctx) error {
	if c.Session != nil {
		if !c.RW.Written() {
			c.Flasher.Out(c.Session)
			c.Session.SessionRelease(c.RW)
		}
	}
	return nil
}

func setsession(c *ctx, key string, value interface{}) error {
	return c.Session.Set(key, value)
}

func startsession(c *ctx, s *session.Manager) error {
	var err error
	c.Session, err = s.SessionStart(c.RW, c.Request)
	if err != nil {
		return err
	}
	return nil
}

var flashfxtension = map[string]interface{}{
	"flasher": flshr,
	"flash":   flash,
}

var FlashFxtension Fxtension = MakeFxtension("flashfxtension", flashfxtension)

type Flashes map[string][]string

type Flasher interface {
	Write(string) []string
	WriteAll() Flashes
	In(session.SessionStore) bool
	Out(session.SessionStore) bool
	Flash(string, string)
}

func NewFlasher() Flasher {
	return &flasher{}
}

type flasher struct {
	readOnce bool
	flashes  Flashes
}

func (f *flasher) Write(key string) []string {
	if ret, ok := f.flashes[key]; ok {
		f.readOnce = true
		return ret
	}
	return nil
}

func (f *flasher) WriteAll() Flashes {
	ret := make(Flashes)
	for k, v := range f.flashes {
		ret[k] = v
	}
	f.readOnce = true
	f.flashes = nil
	return ret
}

func (f *flasher) In(s session.SessionStore) bool {
	if in := s.Get("_flashes"); in != nil {
		if inf, ok := in.(Flashes); ok {
			f.flashes = inf
			return true
		}
	}
	return false
}

func (f *flasher) Out(s session.SessionStore) bool {
	if err := s.Set("_flashes", f.flashes); err != nil {
		return true
	}
	return false
}

func (f *flasher) Flash(key, value string) {
	if f.flashes == nil {
		f.flashes = make(Flashes)
	}
	f.flashes[key] = append(f.flashes[key], value)
}

func Flshr(c Ctx) Flasher {
	fl, _ := c.Call("flasher")
	return fl.(Flasher)
}

func flshr(c *ctx) (Flasher, error) {
	return c.Flasher, nil
}

func flash(c *ctx, category, value string) error {
	c.Flasher.Flash(category, value)
	return nil
}

// MakeCtxFxtension creates a utility Fxtension with miscellaneous functions.
func MakeCtxFxtension(a *App) Fxtension {
	ctxfxtension := map[string]interface{}{
		"env":            envqueryfunc(a),
		"files":          files,
		"get":            getdata,
		"mode":           currentmodefunc(a),
		"out":            out(a),
		"emit":           emit(a),
		"panics":         panics,
		"panicsignal":    panicsignalfunc(a),
		"params":         currentparams,
		"paramString":    paramString,
		"push":           push,
		"rendertemplate": rendertemplatefunc(a),
		"request":        currentrequest,
		"set":            setdata,
		"status":         statusfunc(a),
		"store":          storequeryfunc(a),
		"urlfor":         urlforfunc(a),
	}

	return MakeFxtension("ctxfxtension", ctxfxtension)
}

func envqueryfunc(a *App) func(*ctx, string) interface{} {
	return func(c *ctx, item string) interface{} {
		switch item {
		case "store":
			return a.Env.Store
		case "fxtensions":
			return a.Env.fxtensions
		case "processors":
			return a.Env.ctxprocessors
		default:
			return nil
		}
	}
}

type RequestFiles map[string][]*multipart.FileHeader

func files(c *ctx) RequestFiles {
	if c.Request.MultipartForm.File != nil {
		return c.Request.MultipartForm.File
	}
	return nil
}

var InvalidKey = xrr.NewXrror("Key %s does not exist.").Out

func getdata(c *ctx, key string) (interface{}, error) {
	item, ok := c.Data[key]
	if ok {
		return item, nil
	}
	return nil, InvalidKey(key)
}

func currentmodefunc(a *App) func(c *ctx) *Modes {
	return func(c *ctx) *Modes {
		return a.Mode
	}
}

func currentparams(c *ctx) engine.Params {
	return c.Params
}

func panics(c *ctx) xrr.Xrrors {
	return c.Result.Errors().ByType(xrr.ErrorTypePanic)
}

func panicsignalfunc(a *App) func(*ctx, string) error {
	return func(c *ctx, sig string) error {
		a.Panic(sig)
		return nil
	}
}

func paramString(c *ctx, key string) string {
	for _, v := range c.Params {
		if v.Key == key {
			return v.Value
		}
	}
	return ""
}

func out(a *App) func(*ctx, string) error {
	return func(c *ctx, msg string) error {
		a.Messaging.Out(msg)
		return nil
	}
}

func emit(a *App) func(*ctx, string) error {
	return func(c *ctx, msg string) error {
		a.Messaging.Emit(msg)
		return nil
	}
}

func push(c *ctx, m Manage) error {
	c.push(m)
	return nil
}

func CurrentRequest(c Ctx) *http.Request {
	req, _ := c.Call("request")
	return req.(*http.Request)
}

func currentrequest(c *ctx) *http.Request {
	return c.Request
}

func rendertemplatefunc(a *App) func(*ctx, string, interface{}) error {
	return func(c *ctx, name string, data interface{}) error {
		c.push(func(pc Ctx) {
			td := NewTemplateData(c, data)
			a.Templator.Render(c.RW, name, td)
		})
		return nil
	}
}

func setdata(c *ctx, key string, item interface{}) error {
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.Data[key] = item
	return nil
}

var NoStoreItem = xrr.NewXrror("Could not find StoreItem %s").Out

func storequeryfunc(a *App) func(*ctx, string) (*StoreItem, error) {
	return func(c *ctx, key string) (*StoreItem, error) {
		if item, ok := a.Env.Store[key]; ok {
			return item, nil
		}
		return nil, NoStoreItem(key)
	}
}

// CheckStore is returns a StoreItem and a boolean indicating existence provided
// the current Ctx and a key string.
func CheckStore(c Ctx, key string) (*StoreItem, bool) {
	if item, err := c.Call("store", key); err == nil {
		return item.(*StoreItem), true
	}
	return nil, false
}

var NoUrl = xrr.NewXrror("Unable to get url for route %s with params %s.").Out

func urlforfunc(a *App) func(*ctx, string, bool, []string) (string, error) {
	return func(c *ctx, route string, external bool, params []string) (string, error) {
		if route, ok := a.Routes()[route]; ok {
			routeurl, _ := route.Url(params...)
			if routeurl != nil {
				if external {
					routeurl.Host = c.Request.Host
				}
				return routeurl.String(), nil
			}
		}
		return "", NoUrl(route, params)
	}
}

var readyextensions = []Fxtension{CookieFxtension, FlashFxtension, ResponseFxtension, SessionFxtension}

func BuiltInExtensions(a *App) []Fxtension {
	var ret []Fxtension
	ret = append(ret, readyextensions...)
	ret = append(ret, MakeCtxFxtension(a))
	return ret
}
