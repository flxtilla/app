package flotilla

import (
	"fmt"
	"math"
	"net/http"
	"reflect"
	"time"

	"github.com/thrisp/flotilla/session"
)

type (
	// A Manage function is any function taking a single parameter, *Ctx
	Manage func(*Ctx)

	// Ctx is the primary context for passing & setting data between handlerfunc
	// of a route, constructed from the *App and the app engine context data.
	Ctx struct {
		index      int8
		handlers   []Manage
		deferred   []Manage
		extensions map[string]reflect.Value
		processors map[string]reflect.Value
		rw         responseWriter
		RW         ResponseWriter
		Request    *http.Request
		Session    session.SessionStore
		Data       map[string]interface{}
		App        *App
		Errors     errorMsgs
		*recorder
	}

	recorder struct {
		start     time.Time
		stop      time.Time
		latency   time.Duration
		status    int
		method    string
		path      string
		requester string
	}
)

func (a *App) newCtx() interface{} {
	ctx := &Ctx{index: -1,
		App:        a,
		Data:       make(map[string]interface{}),
		extensions: reflectFuncs(a.extensions),
		processors: reflectFuncs(a.ctxprocessors),
	}
	ctx.RW = &ctx.rw
	return ctx
}

func (a *App) getCtx(rw http.ResponseWriter, req *http.Request, rslt *result) *Ctx {
	ctx := a.p.Get().(*Ctx)
	ctx.recorder = &recorder{}
	if mm, err := a.Env.Store["UPLOAD_SIZE"].Int64(); err == nil {
		req.ParseMultipartForm(mm)
	}
	ctx.Request = req
	ctx.rw.reset(rw)
	for _, p := range rslt.params {
		ctx.Data[p.Key] = p.Value
	}
	ctx.Start()
	return ctx
}

func (a *App) putCtx(ctx *Ctx) {
	ctx.PostProcess(ctx.Request, ctx.RW)
	if !a.Mode.Production {
		a.Send("out", ctx.LogFmt())
	}
	ctx.index = -1
	ctx.Session = nil
	for k, _ := range ctx.Data {
		delete(ctx.Data, k)
	}
	ctx.handlers = nil
	ctx.deferred = nil
	ctx.recorder = nil
	ctx.Errors = nil
	a.p.Put(ctx)
}

func (ctx *Ctx) Start() {
	ctx.StartRecorder()
	ctx.Session, _ = ctx.App.SessionManager.SessionStart(ctx.RW, ctx.Request)
}

func (ctx *Ctx) Release() {
	if !ctx.RW.Written() {
		ctx.Session.SessionRelease(ctx.RW)
	}
}

// Calls a function with name in *Ctx.extensions passing in the given args.
func (ctx *Ctx) Call(name string, args ...interface{}) (interface{}, error) {
	return call(ctx.extensions[name], args...)
}

// Copies the Ctx with handlers set to nil.
func (ctx *Ctx) Copy() *Ctx {
	var rcopy Ctx = *ctx
	rcopy.index = math.MaxInt8 / 2
	rcopy.handlers = nil
	return &rcopy
}

func (ctx *Ctx) events() {
	ctx.Push(func(c *Ctx) { c.Release() })
	ctx.Next()
	for _, fn := range ctx.deferred {
		fn(ctx)
	}
}

// Executes the pending handlers in the chain inside the calling handlectx.
func (ctx *Ctx) Next() {
	ctx.index++
	s := int8(len(ctx.handlers))
	for ; ctx.index < s; ctx.index++ {
		ctx.handlers[ctx.index](ctx)
	}
}

// Push places a handlerfunc in ctx.deferred for execution after all handlerfuncs have run.
func (ctx *Ctx) Push(fn Manage) {
	ctx.deferred = append(ctx.deferred, fn)
}

// Sets a new pair key/value in the current Ctx.
func (ctx *Ctx) Set(key string, item interface{}) {
	ctx.Data[key] = item
}

// Get returns the value for the given key or an error if nonexistent.
func (ctx *Ctx) Get(key string) (interface{}, error) {
	item, ok := ctx.Data[key]
	if ok {
		return item, nil
	}
	return nil, newError("Key %s does not exist.", key)
}

// WriteToHeader writes the specified code and values to the response Head.
// values are 2 string arrays indicating the key first and the value second
// to set in the Head.
func (ctx *Ctx) WriteToHeader(code int, values ...[]string) {
	if code >= 0 {
		ctx.RW.WriteHeader(code)
	}
	ctx.ModifyHeader("set", values...)
}

func (ctx *Ctx) ModifyHeader(action string, values ...[]string) {
	switch action {
	case "set":
		for _, v := range values {
			ctx.RW.Header().Set(v[0], v[1])
		}
	default:
		for _, v := range values {
			ctx.RW.Header().Add(v[0], v[1])
		}
	}
}

func (r *recorder) StartRecorder() {
	r.start = time.Now()
}

func (r *recorder) StopRecorder() {
	r.stop = time.Now()
}

func (r *recorder) Requester(req *http.Request) {
	rqstr := req.Header.Get("X-Real-IP")
	if len(rqstr) == 0 {
		rqstr = req.Header.Get("X-Forwarded-For")
	}
	if len(rqstr) == 0 {
		rqstr = req.RemoteAddr
	}
	r.requester = rqstr
}

func (r *recorder) Latency() time.Duration {
	return r.stop.Sub(r.start)
}

func (r *recorder) PostProcess(req *http.Request, rw ResponseWriter) {
	r.StopRecorder()
	r.latency = r.Latency()
	r.Requester(req)
	r.method = req.Method
	r.path = req.URL.Path
	r.status = rw.Status()
}

func (r *recorder) Fmt() string {
	return fmt.Sprintf("%s	%s	%s	%3d	%s	%s	%s", r.start, r.stop, r.latency, r.status, r.method, r.path, r.requester)
}

func (r *recorder) LogFmt() string {
	return fmt.Sprintf("%v |%s %3d %s| %12v | %s |%s %s %-7s %s",
		r.stop.Format("2006/01/02 - 15:04:05"),
		StatusColor(r.status), r.status, reset,
		r.latency,
		r.requester,
		MethodColor(r.method), reset, r.method,
		r.path)
}
