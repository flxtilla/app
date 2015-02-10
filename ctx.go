package flotilla

import (
	"errors"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/xrr"
)

type Manage func(Ctx)

type Ctx interface {
	Extensor
	Run()
	Next()
	Cancel()
}

var Canceled = errors.New("flotilla.Ctx canceled")

func (a *App) Ctx() func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result, rt *Route) Ctx {
	return func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result, rt *Route) Ctx {
		c := NewCtx(a.extensions, rs)
		c.reset(rq, rw, rt.managers)
		c.Call("start", a.SessionManager)
		return c
	}
}

type context struct {
	parent   *context
	mu       sync.Mutex
	children map[canceler]bool
	done     chan struct{}
	err      error
	value    *ctx
}

type canceler interface {
	cancel(removeFromParent bool, err error)
	Done() <-chan struct{}
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

type handlers struct {
	index    int8
	managers []Manage
	deferred []Manage
}

func defaulthandlers() *handlers {
	return &handlers{index: -1}
}

type ctx struct {
	*engine.Result
	*context
	*handlers
	xrr.Erroror
	Extensor
	rw      responseWriter
	RW      ResponseWriter
	Request *http.Request
	Session session.SessionStore
	Data    map[string]interface{}
}

func emptyCtx() *ctx {
	return &ctx{handlers: defaulthandlers(), Erroror: xrr.DefaultErroror()}
}

func NewCtx(ext map[string]reflect.Value, rs *engine.Result) *ctx {
	c := emptyCtx()
	c.Result = rs
	c.Extensor = newextensor(ext, c)
	c.RW = &c.rw
	return c
}

func (c *ctx) Run() {
	c.push(func(c Ctx) { c.Call("release") })
	c.Next()
	for _, fn := range c.deferred {
		fn(c)
	}
}

func (c *ctx) Next() {
	c.index++
	s := int8(len(c.managers))
	for ; c.index < s; c.index++ {
		c.managers[c.index](c)
	}
}

func (c *ctx) Cancel() {
	c.PostProcess(c.Request, c.RW.Status())
	c.context.cancel(true, Canceled)
}

func (c *ctx) reset(rq *http.Request, rw http.ResponseWriter, m []Manage) {
	c.Request = rq
	c.rw.reset(rw)
	c.context = &context{done: make(chan struct{}), value: c}
	c.handlers = defaulthandlers()
	c.managers = m
}

func (c *ctx) replicate() *ctx {
	child := &context{parent: c.context, done: make(chan struct{}), value: c}
	propagateCancel(c.context, child)
	var rcopy ctx = *c
	rcopy.context = child
	return &rcopy
}

func (c *ctx) rerun(managers ...Manage) {
	if c.index != -1 {
		c.index = -1
	}
	c.managers = managers
	// calling Run() here would be preferable, but causes a double render with an undetermined cause; its a bug
	c.index++
	x := int8(len(c.managers))
	for ; c.index < x; c.index++ {
		c.managers[c.index](c)
	}
}

func (c *ctx) push(fn Manage) {
	c.deferred = append(c.deferred, fn)
}

func CheckStore(c Ctx, key string) (*StoreItem, bool) {
	if item, err := c.Call("store", c, key); err == nil {
		return item.(*StoreItem), true
	}
	return nil, false
}
