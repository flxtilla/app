package flotilla

import (
	"mime/multipart"
	"net/http"
	"net/url"
)

var (
	builtinextensions = map[string]interface{}{
		"abort":            abort,
		"allflashmessages": allflashmessages,
		"cookie":           cookie,
		"cookies":          cookies,
		"files":            files,
		"flash":            flash,
		"flashmessages":    flashmessages,
		"form":             form,
		"redirect":         redirect,
		"rendertemplate":   rendertemplate,
		"serveplain":       serveplain,
		"servefile":        servefile,
		"status":           status,
		"urlfor":           urlfor,
	}
)

type (
	RequestFiles map[string][]*multipart.FileHeader
)

func abort(c *Ctx, code int) error {
	if code >= 0 {
		c.RW.WriteHeader(code)
	}
	return nil
}

func (ctx *Ctx) Abort(code int) {
	ctx.Call("abort", ctx, code)
}

func status(c *Ctx, code int) error {
	rslt := c.App.engine.status(code)
	rslt.manage(c)
	return nil
}

func (ctx *Ctx) Status(code int) {
	ctx.Call("status", ctx, code)
}

func redirect(ctx *Ctx, code int, location string) error {
	if code >= 300 && code <= 308 {
		ctx.Push(func(c *Ctx) {
			http.Redirect(c.RW, c.Request, location, code)
			c.RW.WriteHeaderNow()
		})
		return nil
	} else {
		return newError("Cannot send a redirect with status code %d", code)
	}
}

func (ctx *Ctx) Redirect(code int, location string) {
	ctx.Call("redirect", ctx, code, location)
}

func serveplain(ctx *Ctx, code int, data []byte) error {
	ctx.Push(func(c *Ctx) {
		c.WriteToHeader(code, []string{"Content-Type", "text/plain"})
		c.RW.Write(data)
	})
	return nil
}

func (ctx *Ctx) ServePlain(code int, data []byte) {
	ctx.Call("serveplain", ctx, code, data)
}

func servefile(ctx *Ctx, f http.File) error {
	fi, err := f.Stat()
	if err == nil {
		http.ServeContent(ctx.RW, ctx.Request, fi.Name(), fi.ModTime(), f)
	}
	return err
}

func (ctx *Ctx) ServeFile(f http.File) {
	ctx.Call("servefile", ctx, f)
}

func rendertemplate(ctx *Ctx, name string, data interface{}) error {
	//td := TemplateData(ctx, data)
	//ctx.Push(func(c *Ctx) {
	//	c.App.Templator.Render(c.RW, name, td)
	//})
	return nil
}

func (ctx *Ctx) RenderTemplate(name string, data interface{}) {
	ctx.Call("rendertemplate", ctx, name, data)
}

func urlfor(ctx *Ctx, route string, external bool, params []string) (string, error) {
	//if route, ok := ctx.App.Routes()[route]; ok {
	//	routeurl, _ := route.Url(params...)
	//	if routeurl != nil {
	//		if external {
	//			routeurl.Host = ctx.Request.Host
	//		}
	//		return routeurl.String(), nil
	//	}
	//}
	return "", newError("unable to get url for route %s with params %s", route, params)
}

func (ctx *Ctx) UrlRelative(route string, params ...string) string {
	ret, err := ctx.Call("urlfor", ctx, route, false, params)
	if err != nil {
		return err.Error()
	}
	return ret.(string)
}

func (ctx *Ctx) UrlExternal(route string, params ...string) string {
	ret, err := ctx.Call("urlfor", ctx, route, true, params)
	if err != nil {
		return err.Error()
	}
	return ret.(string)
}

func flash(ctx *Ctx, category string, message string) error {
	if fl := ctx.Session.Get("_flashes"); fl != nil {
		if fls, ok := fl.(map[string]string); ok {
			fls[category] = message
			ctx.Session.Set("_flashes", fls)
		}
	} else {
		fl := make(map[string]string)
		fl[category] = message
		ctx.Session.Set("_flashes", fl)
	}
	return nil
}

func (ctx *Ctx) Flash(category string, message string) {
	ctx.Call("flash", ctx, category, message)
}

func flashmessages(ctx *Ctx, categories []string) []string {
	var ret []string
	if fl := ctx.Session.Get("_flashes"); fl != nil {
		if fls, ok := fl.(map[string]string); ok {
			for k, v := range fls {
				if existsIn(k, categories) {
					ret = append(ret, v)
					delete(fls, k)
				}
			}
			ctx.Session.Set("_flashes", fls)
		}
	}
	return ret
}

func (ctx *Ctx) FlashMessages(categories ...string) []string {
	ret, _ := ctx.Call("flashmessages", ctx, categories)
	return ret.([]string)
}

func allflashmessages(ctx *Ctx) map[string]string {
	var ret map[string]string
	if fl := ctx.Session.Get("_flashes"); fl != nil {
		if fls, ok := fl.(map[string]string); ok {
			ret = fls
		}
	}
	ctx.Session.Delete("_flashes")
	return ret
}

func (ctx *Ctx) AllFlashMessages() map[string]string {
	ret, _ := ctx.Call("allflashmessages", ctx)
	return ret.(map[string]string)
}

func form(ctx *Ctx) url.Values {
	return ctx.Request.Form
}

func (ctx *Ctx) Form() (url.Values, error) {
	ret, err := ctx.Call("form", ctx)
	return ret.(url.Values), err
}

func files(ctx *Ctx) RequestFiles {
	if ctx.Request.MultipartForm.File != nil {
		return ctx.Request.MultipartForm.File
	}
	return nil
}

func (ctx *Ctx) Files() (RequestFiles, error) {
	ret, err := ctx.Call("files", ctx)
	return ret.(RequestFiles), err
}
