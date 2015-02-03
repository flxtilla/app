package flotilla

import (
	"fmt"
	"html/template"
	"reflect"
)

type (
	TData map[string]interface{}
)

func templateData(any interface{}) TData {
	if rcvd, ok := any.(map[string]interface{}); ok {
		td := rcvd
		return td
	} else {
		td := make(map[string]interface{})
		td["Any"] = any
		return td
	}
}

func TemplateData(c *Ctx, any interface{}) TData {
	td := templateData(any)
	td["Ctx"] = c.Copy()
	td["Request"] = c.Request
	td["Session"] = c.Session
	for k, v := range c.Data {
		td[k] = v
	}
	td["Flash"] = c.AllFlashMessages()
	td.contextProcessors(c)
	return td
}

func (t TData) GetFlashMessages(categories ...string) []string {
	var ret []string
	if fls, ok := t["Flash"].(map[string]string); ok {
		for k, v := range fls {
			if existsIn(k, categories) {
				ret = append(ret, v)
			}
		}
	}
	return ret
}

func (t TData) UrlFor(route string, external bool, params ...string) string {
	if ctx, ok := t["Ctx"].(*Ctx); ok {
		ret, err := ctx.Call("urlfor", ctx, route, external, params)
		if err != nil {
			return newError(fmt.Sprint("%s", err)).Error()
		}
		return ret.(string)
	}
	return fmt.Sprintf("Unable to return a url from: %s, %s, external(%t)", route, params, external)
}

func (t TData) HTML(name string) template.HTML {
	if fn, ok := t.ctxPrc(name); ok {
		res, err := call(fn)
		if err != nil {
			return template.HTML(err.Error())
		}
		if ret, ok := res.(template.HTML); ok {
			return ret
		}
	}
	return template.HTML(fmt.Sprintf("<p>context processor %s unprocessable by HTML</p>", name))
}

func (t TData) STRING(name string) string {
	if fn, ok := t.ctxPrc(name); ok {
		res, err := call(fn)
		if err != nil {
			return err.Error()
		}
		if ret, ok := res.(string); ok {
			return ret
		}
	}
	return fmt.Sprintf("context processor %s unprocessable by STRING", name)
}

func (t TData) CALL(name string) interface{} {
	if fn, ok := t.ctxPrc(name); ok {
		if res, err := call(fn); err == nil {
			return res
		} else {
			return err
		}
	}
	return fmt.Sprintf("context processor %s cannot be processed by CALL", name)
}

func (t TData) ctxPrc(name string) (reflect.Value, bool) {
	if fn, ok := t[name]; ok {
		if fn, ok := fn.(reflect.Value); ok {
			return fn, true
		}
	}
	return reflect.Value{}, false
}

func (t TData) contextProcessor(fn reflect.Value, c *Ctx) reflect.Value {
	newfn := func() (interface{}, error) {
		return call(fn, c)
	}
	return valueFunc(newfn)
}

func (t TData) contextProcessors(c *Ctx) {
	for k, fn := range c.App.ctxprocessors {
		t[k] = t.contextProcessor(fn, c)
	}
}
