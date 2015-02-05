package flotilla

import (
	"fmt"
	"html/template"
	"reflect"
)

type (
	TemplateData map[string]interface{}
)

func baseTemplateData(any interface{}) TemplateData {
	if rcvd, ok := any.(map[string]interface{}); ok {
		return rcvd
	} else {
		td := make(map[string]interface{})
		td["Any"] = any
		return td
	}
}

func NewTemplateData(c *Ctx, any interface{}) TemplateData {
	t := baseTemplateData(any)
	t["Ctx"] = c.Copy()
	t["Request"] = c.Request
	t["Session"] = c.Session
	for k, v := range c.Data {
		t[k] = v
	}
	t["Flash"] = getallflash(c)
	t.setCtxProcessors(c)
	return t
}

func getallflash(c *Ctx) map[string]string {
	ret, _ := c.Call("flashed")
	return ret.(map[string]string)
}

func (t TemplateData) Flashes(categories ...string) []string {
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

func (t TemplateData) UrlFor(route string, external bool, params ...string) string {
	if ctx, ok := t["Ctx"].(*Ctx); ok {
		ret, err := ctx.Call("urlfor", route, external, params)
		if err != nil {
			return newError(fmt.Sprint("%s", err)).Error()
		}
		return ret.(string)
	}
	return fmt.Sprintf("Unable to return a url from: %s, %s, external(%t)", route, params, external)
}

func (t TemplateData) HTML(name string) template.HTML {
	if fn, ok := t.getCtxProcessor(name); ok {
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

func (t TemplateData) STRING(name string) string {
	if fn, ok := t.getCtxProcessor(name); ok {
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

func (t TemplateData) CALL(name string) interface{} {
	if fn, ok := t.getCtxProcessor(name); ok {
		if res, err := call(fn); err == nil {
			return res
		} else {
			return err
		}
	}
	return fmt.Sprintf("context processor %s cannot be processed by CALL", name)
}

func (t TemplateData) getCtxProcessor(name string) (reflect.Value, bool) {
	if fn, ok := t[name]; ok {
		if fn, ok := fn.(reflect.Value); ok {
			return fn, true
		}
	}
	return reflect.Value{}, false
}

func (t TemplateData) setCtxProcessor(fn reflect.Value, c *Ctx) reflect.Value {
	newfn := func() (interface{}, error) {
		return call(fn, c)
	}
	return valueFunc(newfn)
}

func processorsFromEnv(c *Ctx) map[string]reflect.Value {
	ret, err := c.Call("env", "processors")
	if err == nil {
		return ret.(map[string]reflect.Value)
	}
	return nil
}

func (t TemplateData) setCtxProcessors(c *Ctx) {
	for k, fn := range processorsFromEnv(c) {
		t[k] = t.setCtxProcessor(fn, c)
	}
}
