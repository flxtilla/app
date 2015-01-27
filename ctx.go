package flotilla

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/thrisp/flotilla/session"
)

type (
	Manage func(*Ctx)

	canceler interface {
		cancel(removeFromParent bool, err error)
		Done() <-chan struct{}
	}

	context struct {
		mu       sync.Mutex
		parent   *context
		children map[canceler]bool
		done     chan struct{}
		err      error
		value    *Ctx
	}

	CancelFunc func(*App, *Ctx)

	handlers struct {
		index    int8
		managers []Manage
		deferred []Manage
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

	Ctx struct {
		lookup     *result
		extensions map[string]reflect.Value
		rw         responseWriter
		RW         ResponseWriter
		Request    *http.Request
		Session    session.SessionStore
		Data       map[string]interface{}
		App        *App
		Errors     errorMsgs
		*context
		*handlers
		*recorder
	}
)

var (
	Canceled = errors.New("flotilla.Ctx canceled")
)

func (a *App) newCtx() interface{} {
	ctx := &Ctx{
		handlers:   defaulthandlers(),
		App:        a,
		extensions: a.extensions,
	}
	ctx.RW = &ctx.rw
	return ctx
}

func (a *App) getCtx(rw http.ResponseWriter, req *http.Request) (*Ctx, CancelFunc) {
	ctx := a.getctx(rw, req)
	cancel := func(app *App, c *Ctx) { c.context.cancel(true, Canceled); app.putCtx(c) }
	return ctx, cancel
}

func (a *App) getctx(rw http.ResponseWriter, req *http.Request) *Ctx {
	ctx := a.p.Get().(*Ctx)
	ctx.context = &context{done: make(chan struct{}), value: ctx}
	ctx.recorder = NewRecorder()
	if mm, err := a.Env.Store["UPLOAD_SIZE"].Int64(); err == nil {
		req.ParseMultipartForm(mm)
	}
	ctx.Request = req
	ctx.rw.reset(rw)
	ctx.Start()
	return ctx
}

func (a *App) putCtx(ctx *Ctx) {
	ctx.PostProcess(ctx.Request, ctx.RW)
	if !a.Mode.Production {
		a.Send("out", ctx.LogFmt())
	}
	ctx.handlers = defaulthandlers()
	ctx.Session = nil
	ctx.Data = nil
	ctx.managers = nil
	ctx.deferred = nil
	ctx.recorder = nil
	ctx.Errors = nil
	ctx.context = nil
	a.p.Put(ctx)
}

func Copy(c *Ctx) *Ctx {
	child := &context{parent: c.context, done: make(chan struct{}), value: c}
	propagateCancel(c.context, child)
	var rcopy Ctx = *c
	rcopy.context = child
	return &rcopy
}

func propagateCancel(p *context, child canceler) {
	if p.Done() == nil {
		return
	}
	p.mu.Lock()
	if p.err != nil {
		// parent has already been canceled
		child.cancel(false, p.err)
	} else {
		if p.children == nil {
			p.children = make(map[canceler]bool)
		}
		p.children[child] = true
	}
	p.mu.Unlock()
}

func (c *context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *context) Done() <-chan struct{} {
	return c.done
}

func (c *context) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *context) Value(key interface{}) interface{} {
	return c.value
}

func (c *context) cancel(removeFromParent bool, err error) {
	if err == nil {
		panic("Ctx.context: internal error: missing cancel error")
	}
	c.mu.Lock()
	c.err = err
	close(c.done)
	for child := range c.children {
		child.cancel(false, err)
	}
	c.children = nil
	c.mu.Unlock()

	if removeFromParent {
		if c.children != nil {
			delete(c.children, c)
		}
	}
}

func defaulthandlers() *handlers {
	return &handlers{index: -1}
}

func NewRecorder() *recorder {
	r := &recorder{}
	r.StartRecorder()
	return r
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

func (ctx *Ctx) Result(r *result) {
	ctx.lookup = r
	for _, p := range r.params {
		ctx.Set(p.Key, p.Value)
	}
}

func (ctx *Ctx) Start() {
	ctx.Session, _ = ctx.App.SessionManager.SessionStart(ctx.RW, ctx.Request)
}

func (ctx *Ctx) Release() {
	if !ctx.RW.Written() {
		ctx.Session.SessionRelease(ctx.RW)
	}
}

func (ctx *Ctx) Call(name string, args ...interface{}) (interface{}, error) {
	return call(ctx.extensions[name], args...)
}

func (ctx *Ctx) events() {
	ctx.Push(func(c *Ctx) { c.Release() })
	ctx.Next()
	for _, fn := range ctx.deferred {
		fn(ctx)
	}
}

func (ctx *Ctx) Next() {
	ctx.index++
	s := int8(len(ctx.managers))
	for ; ctx.index < s; ctx.index++ {
		ctx.managers[ctx.index](ctx)
	}
}

func (ctx *Ctx) Push(fn Manage) {
	ctx.deferred = append(ctx.deferred, fn)
}

func (ctx *Ctx) Set(key string, item interface{}) {
	if ctx.Data == nil {
		ctx.Data = make(map[string]interface{})
	}
	ctx.Data[key] = item
}

func (ctx *Ctx) Get(key string) (interface{}, error) {
	item, ok := ctx.Data[key]
	if ok {
		return item, nil
	}
	return nil, newError("Key %s does not exist.", key)
}

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
