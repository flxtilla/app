package template

import (
	"fmt"
	"html/template"
	"reflect"
)

type TemplateData interface {
	//UrlFor(string, bool, ...string) string
	HTML(string) template.HTML
	STRING(string) string
	CALL(string) interface{}
}

type templateData map[string]interface{}

func NewTemplateData(in interface{}) TemplateData {
	ret := make(templateData)
	if rcvd, ok := in.(map[string]interface{}); ok {
		for k, v := range rcvd {
			ret[k] = v
		}
	} else {
		ret["Any"] = in
	}
	return ret
}

//func NewTemplateData(c Ctx, any interface{}) TemplateData {
//t := baseTemplateData(any)
//t["Ctx"] = c.replicate()
//t["Request"] = c.Request
//t["Session"] = c.Session
//for k, v := range c.Data {
//	t[k] = v
//}
//t["Flash"] = c.Flasher
//t.setCtxProcessors(c)
//return t
//}

//func (t templateData) UrlFor(route string, external bool, params ...string) string {
//if c, ok := t["Ctx"].(Ctx); ok {
//	ret, err := c.Call("urlfor", route, external, params)
//	if err != nil {
//		return err.Error()
//	}
//	return ret.(string)
//}
//	return fmt.Sprintf("Unable to return a url from: %s, %s, external(%t)", route, params, external)
//}

func (t templateData) HTML(name string) template.HTML {
	if fn, ok := t.getCtxProcessor(name); ok {
		res, err := call(fn)
		if err != nil {
			return template.HTML(err.Error())
		}
		if result, ok := res.(template.HTML); ok {
			return result
		}
	}
	return template.HTML(fmt.Sprintf("<p>context processor %s unprocessable by HTML</p>", name))
}

func (t templateData) STRING(name string) string {
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

func (t templateData) CALL(name string) interface{} {
	if fn, ok := t.getCtxProcessor(name); ok {
		if result, err := call(fn); err != nil {
			return err.Error()
		} else {
			return result
		}
	}
	return fmt.Sprintf("context processor %s cannot be processed by CALL", name)
}

func (t templateData) getCtxProcessor(name string) (reflect.Value, bool) {
	if fn, ok := t[name]; ok {
		if fn, ok := fn.(reflect.Value); ok {
			return fn, true
		}
	}
	return reflect.Value{}, false
}

//func (t templateData) setCtxProcessor(fn reflect.Value, c *ctx) reflect.Value {
//	newfn := func() (interface{}, error) {
//		return call(fn, c)
//	}
//	return valueFunc(newfn)
//}

//func processorsFromEnv(c *ctx) map[string]reflect.Value {
//	ret, err := c.Call("env", "processors")
//	if err == nil {
//		return ret.(map[string]reflect.Value)
//	}
//	return nil
//}

//func (t templateData) setCtxProcessors(c *ctx) {
//	for k, fn := range processorsFromEnv(c) {
//		t[k] = t.setCtxProcessor(fn, c)
//	}
//}
