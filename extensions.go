package flotilla

import (
	"mime/multipart"
	"net/http"
	"reflect"

	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/xrr"
)

type (
	// The Extensor interface provides access to varying functions.
	Extensor interface {
		Call(string, ...interface{}) (interface{}, error)
	}

	extensor struct {
		ext map[string]reflect.Value
		ctx *ctx
	}

	RequestFiles map[string][]*multipart.FileHeader
)

func newextensor(e map[string]reflect.Value, c *ctx) Extensor {
	return &extensor{ext: e, ctx: c}
}

func (e *extensor) Call(name string, args ...interface{}) (interface{}, error) {
	var ctxargs []interface{}
	ctxargs = append(ctxargs, e.ctx)
	ctxargs = append(ctxargs, args...)
	return call(e.ext[name], ctxargs...)
}

func BuiltInExtensions(a *App) map[string]interface{} {
	b := map[string]interface{}{
		"abort":           abort,
		"cookie":          cookie,
		"cookies":         cookies,
		"deletesession":   deletesession,
		"env":             envqueryfunc(a),
		"files":           files,
		"flash":           flash,
		"flashes":         flashes,
		"flashed":         flashed,
		"get":             getdata,
		"getsession":      getsession,
		"headernow":       headernow,
		"headerwrite":     headerwrite,
		"headermodify":    headermodify,
		"iswritten":       iswritten,
		"mode":            currentmodefunc(a),
		"panics":          panics,
		"panicsignal":     panicsignalfunc(a),
		"push":            push,
		"readcoookies":    readcookies,
		"redirect":        redirect,
		"request":         currentrequest,
		"release":         releasesession,
		"rendertemplate":  rendertemplatefunc(a),
		"securecookie":    securecookie,
		"serveplain":      serveplain,
		"servefile":       servefile,
		"set":             setdata,
		"session":         returnsession,
		"setsession":      setsession,
		"start":           startsession,
		"status":          statusfunc(a),
		"store":           storequeryfunc(a),
		"writetoresponse": writetoresponse,
		"urlfor":          urlforfunc(a),
	}
	return b
}

func abort(c *ctx, code int) error {
	if code >= 0 {
		c.RW.WriteHeader(code)
	}
	return nil
}

func redirect(c *ctx, code int, location string) error {
	if code >= 300 && code <= 308 {
		c.push(func(pc Ctx) {
			http.Redirect(c.RW, c.Request, location, code)
			c.RW.WriteHeaderNow()
		})
		return nil
	} else {
		return xrr.NewError("Cannot send a redirect with status code %d", code)
	}
}

func writetoresponse(c *ctx, data []byte) error {
	c.RW.Write(data)
	return nil
}

func serveplain(c *ctx, code int, data []byte) error {
	c.push(func(pc Ctx) {
		headerwrite(c, code, []string{"Content-Type", "text/plain"})
		c.RW.Write(data)
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

func rendertemplatefunc(a *App) func(*ctx, string, interface{}) error {
	return func(c *ctx, name string, data interface{}) error {
		c.push(func(pc Ctx) {
			td := NewTemplateData(c, data)
			a.Templator.Render(c.RW, name, td)
		})
		return nil
	}
}

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
		return "", xrr.NewError("unable to get url for route %s with params %s", route, params)
	}
}

func flash(c *ctx, category string, message string) error {
	if fl := c.Session.Get("_flashes"); fl != nil {
		if fls, ok := fl.(map[string][]string); ok {
			fls[category] = append(fls[category], message)
			c.Session.Set("_flashes", fls)
		}
	} else {
		fl := make(map[string][]string)
		fl[category] = append(fl[category], message)
		c.Session.Set("_flashes", fl)
	}
	return nil
}

func flashes(c *ctx, categories []string) []string {
	var ret []string
	if fl := c.Session.Get("_flashes"); fl != nil {
		if fls, ok := fl.(map[string][]string); ok {
			for k, v := range fls {
				if existsIn(k, categories) {
					ret = v //append(ret, v)
					delete(fls, k)
				}
			}
			c.Session.Set("_flashes", fls)
		}
	}
	return ret
}

func flashed(c *ctx) map[string][]string {
	var ret map[string][]string
	if fl := c.Session.Get("_flashes"); fl != nil {
		if fls, ok := fl.(map[string][]string); ok {
			ret = fls
		}
	}
	c.Session.Delete("_flashes")
	return ret
}

//func form(c *ctx) url.Values {
//c.Request.ParseMultipartForm(e.app.Env.Store["UPLOAD_SIZE"].Int64())
//return c.Request.Form
//}

func files(c *ctx) RequestFiles {
	if c.Request.MultipartForm.File != nil {
		return c.Request.MultipartForm.File
	}
	return nil
}

func setdata(c *ctx, key string, item interface{}) error {
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.Data[key] = item
	return nil
}

func getdata(c *ctx, key string) (interface{}, error) {
	item, ok := c.Data[key]
	if ok {
		return item, nil
	}
	return nil, xrr.NewError("Key %s does not exist.", key)
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

func envqueryfunc(a *App) func(*ctx, string) interface{} {
	return func(c *ctx, item string) interface{} {
		switch item {
		case "store":
			return a.Env.Store
		case "extensions":
			return a.Env.extensions
		case "processors":
			return a.Env.ctxprocessors
		default:
			return nil
		}
	}
}

func storequeryfunc(a *App) func(*ctx, string) (*StoreItem, error) {
	return func(c *ctx, key string) (*StoreItem, error) {
		if item, ok := a.Env.Store[key]; ok {
			return item, nil
		}
		return nil, xrr.NewError("Could not find StoreItem")
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

func startsession(c *ctx, s *session.Manager) error {
	var err error
	c.Session, err = s.SessionStart(c.RW, c.Request)
	if err != nil {
		return err
	}
	return nil
}

func returnsession(c *ctx) session.SessionStore {
	return c.Session
}

func getsession(c *ctx, key string) interface{} {
	return c.Session.Get(key)
}

func setsession(c *ctx, key string, value interface{}) error {
	return c.Session.Set(key, value)
}

func deletesession(c *ctx, key string) error {
	return c.Session.Delete(key)
}

func releasesession(c *ctx) error {
	if c.Session != nil {
		if !c.RW.Written() {
			c.Session.SessionRelease(c.RW)
		}
	}
	return nil
}

func iswritten(c *ctx) bool {
	return c.RW.Written()
}

func push(c *ctx, m Manage) error {
	c.push(m)
	return nil
}

func panics(c *ctx) xrr.ErrorMsgs {
	return c.Result.Errors().ByType(xrr.ErrorTypePanic)
	// ? combine with c.Errors().ByType(xrr.ErrorTypePanic)
}

func panicsignalfunc(a *App) func(c *ctx, s string) error {
	return func(c *ctx, sig string) error {
		a.Panic(sig)
		return nil
	}
}

func currentrequest(c *ctx) *http.Request {
	return c.Request
}

func currentmodefunc(a *App) func(c *ctx) *Modes {
	return func(c *ctx) *Modes {
		return a.Mode
	}
}
