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
	canceler interface {
		cancel(removeFromParent bool, err error)
		Done() <-chan struct{}
	}

	context struct {
		parent   *context
		mu       sync.Mutex
		children map[canceler]bool
		done     chan struct{}
		err      error
		value    *Ctx
	}

	CancelFunc Manage

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
		App        *App
		rw         responseWriter
		RW         ResponseWriter
		Request    *http.Request
		Session    session.SessionStore
		extensions map[string]reflect.Value
		Data       map[string]interface{}
		Errors     errorMsgs
		*context
		*handlers
		*recorder
	}
)

var (
	Canceled = errors.New("flotilla.Ctx canceled")
)

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

func NewCtx(e *engine) *Ctx {
	c := &Ctx{handlers: defaulthandlers()}
	c.App = e.app
	c.extensions = e.app.extensions
	c.RW = &c.rw
	return c
}

func (c *Ctx) Reset(req *http.Request, rw http.ResponseWriter) {
	c.Request = req
	c.rw.reset(rw)
	c.context = &context{done: make(chan struct{}), value: c}
	c.handlers = defaulthandlers()
	c.recorder = NewRecorder()
}

func (c *Ctx) Result(r *result) {
	for _, p := range r.params {
		c.Set(p.Key, p.Value)
	}
}

func (c *Ctx) Start(s *session.Manager) {
	c.Session, _ = s.SessionStart(c.RW, c.Request)
}

func (c *Ctx) Copy() *Ctx {
	child := &context{parent: c.context, done: make(chan struct{}), value: c}
	propagateCancel(c.context, child)
	var rcopy Ctx = *c
	rcopy.context = child
	return &rcopy
}

func (ctx *Ctx) Run(managers ...Manage) {
	ctx.managers = managers
	ctx.Push(func(c *Ctx) { c.Release() })
	ctx.Next()
	for _, fn := range ctx.deferred {
		fn(ctx)
	}
}

func (ctx *Ctx) ReRun(managers ...Manage) {
	if ctx.index != -1 {
		ctx.index = -1
	}
	ctx.Run(managers...)
}

func (ctx *Ctx) Next() {
	ctx.index++
	s := int8(len(ctx.managers))
	for ; ctx.index < s; ctx.index++ {
		ctx.managers[ctx.index](ctx)
	}
}

func (c *Ctx) Cancel() {
	c.context.cancel(true, Canceled)
	c.recorder.PostProcess(c.Request, c.RW)
	if !c.App.Mode.Production {
		c.App.Send("out", c.LogFmt())
	}
	c.handlers = defaulthandlers()
	c.Session = nil
	c.Data = nil
	c.recorder = nil
	c.Errors = nil
	c.context = nil
}

func (ctx *Ctx) Release() {
	if !ctx.RW.Written() {
		ctx.Session.SessionRelease(ctx.RW)
	}
}

func (ctx *Ctx) Call(name string, args ...interface{}) (interface{}, error) {
	return call(ctx.extensions[name], args...)
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
