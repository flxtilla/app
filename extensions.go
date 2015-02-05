package flotilla

import (
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
)

type (
	Extensor interface {
		Call(string, ...interface{}) (interface{}, error)
	}

	extensor struct {
		ext map[string]reflect.Value
		ctx *Ctx
	}

	RequestFiles map[string][]*multipart.FileHeader
)

func newextensor(e map[string]reflect.Value, c *Ctx) Extensor {
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
		"abort":          abort,
		"cookie":         cookie,
		"cookies":        cookies,
		"env":            envquery(a),
		"files":          files,
		"flash":          flash,
		"flashes":        flashes,
		"flashed":        flashed,
		"form":           form,
		"get":            getdata,
		"headerwrite":    headerwrite,
		"headermodify":   headermodify,
		"redirect":       redirect,
		"rendertemplate": rendertemplatefunc(a),
		"serveplain":     serveplain,
		"servefile":      servefile,
		"set":            setdata,
		"status":         a.engine.statusfunc,
		"store":          storequery(a),
		"urlfor":         urlforfunc(a),
	}
	return b
}

func abort(c *Ctx, code int) error {
	if code >= 0 {
		c.RW.WriteHeader(code)
	}
	return nil
}

func redirect(c *Ctx, code int, location string) error {
	if code >= 300 && code <= 308 {
		c.Push(func(c *Ctx) {
			http.Redirect(c.RW, c.Request, location, code)
			c.RW.WriteHeaderNow()
		})
		return nil
	} else {
		return newError("Cannot send a redirect with status code %d", code)
	}
}

func serveplain(c *Ctx, code int, data []byte) error {
	c.Push(func(c *Ctx) {
		headerwrite(c, code, []string{"Content-Type", "text/plain"})
		c.RW.Write(data)
	})
	return nil
}

func servefile(c *Ctx, f http.File) error {
	fi, err := f.Stat()
	if err == nil {
		http.ServeContent(c.RW, c.Request, fi.Name(), fi.ModTime(), f)
	}
	return err
}

func rendertemplatefunc(a *App) func(*Ctx, string, interface{}) error {
	return func(c *Ctx, name string, data interface{}) error {
		td := NewTemplateData(c, data)
		c.Push(func(pc *Ctx) {
			a.Templator.Render(pc.RW, name, td)
		})
		return nil
	}
}

func urlforfunc(a *App) func(*Ctx, string, bool, []string) (string, error) {
	return func(c *Ctx, route string, external bool, params []string) (string, error) {
		if route, ok := a.Routes()[route]; ok {
			routeurl, _ := route.Url(params...)
			if routeurl != nil {
				if external {
					routeurl.Host = c.Request.Host
				}
				return routeurl.String(), nil
			}
		}
		return "", newError("unable to get url for route %s with params %s", route, params)
	}
}

func flash(c *Ctx, category string, message string) error {
	if fl := c.Session.Get("_flashes"); fl != nil {
		if fls, ok := fl.(map[string]string); ok {
			fls[category] = message
			c.Session.Set("_flashes", fls)
		}
	} else {
		fl := make(map[string]string)
		fl[category] = message
		c.Session.Set("_flashes", fl)
	}
	return nil
}

func flashes(c *Ctx, categories []string) []string {
	var ret []string
	if fl := c.Session.Get("_flashes"); fl != nil {
		if fls, ok := fl.(map[string]string); ok {
			for k, v := range fls {
				if existsIn(k, categories) {
					ret = append(ret, v)
					delete(fls, k)
				}
			}
			c.Session.Set("_flashes", fls)
		}
	}
	return ret
}

func flashed(c *Ctx) map[string]string {
	var ret map[string]string
	if fl := c.Session.Get("_flashes"); fl != nil {
		if fls, ok := fl.(map[string]string); ok {
			ret = fls
		}
	}
	c.Session.Delete("_flashes")
	return ret
}

func form(c *Ctx) url.Values {
	return c.Request.Form
}

func files(c *Ctx) RequestFiles {
	if c.Request.MultipartForm.File != nil {
		return c.Request.MultipartForm.File
	}
	return nil
}

func setdata(c *Ctx, key string, item interface{}) error {
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.Data[key] = item
	return nil
}

func getdata(c *Ctx, key string) (interface{}, error) {
	item, ok := c.Data[key]
	if ok {
		return item, nil
	}
	return nil, newError("Key %s does not exist.", key)
}

func headerwrite(c *Ctx, code int, values ...[]string) error {
	if code >= 0 {
		c.RW.WriteHeader(code)
	}
	headermodify(c, "set", values...)
	return nil
}

func headermodify(c *Ctx, action string, values ...[]string) error {
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

func envquery(a *App) func(*Ctx, string) interface{} {
	return func(c *Ctx, item string) interface{} {
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

func storequery(a *App) func(*Ctx, string) (*StoreItem, error) {
	return func(c *Ctx, key string) (*StoreItem, error) {
		if item, ok := a.Env.Store[key]; ok {
			return item, nil
		}
		return nil, newError("Could not find StoreItem")
	}
}
