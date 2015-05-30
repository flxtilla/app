package flotilla

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thrisp/flotilla/engine"
)

type Routes map[string]*Route

// Routes returns a map of all routes attached to the app.
func (a *App) Routes() Routes {
	allroutes := make(Routes)
	for _, blueprint := range a.Blueprints() {
		for _, route := range blueprint.Routes {
			allroutes[route.Name()] = route
		}
	}
	return allroutes
}

type MakeCtxFunc func(w http.ResponseWriter, rq *http.Request, rs *engine.Result, rt *Route) Ctx

type Route struct {
	name       string
	Registered bool
	Blueprint  *Blueprint
	Static     bool
	Method     string
	Base       string
	Path       string
	Managers   []Manage
	MakeCtx    MakeCtxFunc
}

func (rt *Route) rule(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
	c := rt.MakeCtx(rw, rq, rs, rt)
	c.Run()
	c.Cancel()
}

// NewRoute returns a new, non-static, route instance with the given method, path,and Manage functions.
//func NewRoute(method string, path string, managers []Manage) *Route {
func NewRoute(conf ...RouteConf) *Route {
	rt := &Route{}
	err := rt.Configure(conf...)
	if err != nil {
		panic(fmt.Sprintf("[FLOTILLA] route configuration error: %s", err.Error()))
	}
	return rt
}

type RouteConf func(*Route) error

func defaultRouteConf(method string, path string, managers []Manage) RouteConf {
	return func(rt *Route) error {
		rt.Method = method
		rt.Base = path
		rt.Managers = managers
		return nil
	}
}

func staticRouteConf(method string, path string, managers []Manage) RouteConf {
	return func(rt *Route) error {
		rt.Method = method
		rt.Static = true
		if fp := strings.Split(path, "/"); fp[len(fp)-1] != "*filepath" {
			rt.Base = filepath.ToSlash(filepath.Join(path, "/*filepath"))
		} else {
			rt.Base = path
		}
		rt.Managers = managers
		return nil
	}
}
func (rt *Route) Configure(conf ...RouteConf) error {
	var err error
	for _, c := range conf {
		err = c(rt)
	}
	return err
}

// Name returns the route name.
func (rt *Route) Name() string {
	if rt.name == "" {
		return Named(rt)
	}
	return rt.name
}

func (rt *Route) Rename(name string) {
	rt.name = name
}

func Named(rt *Route) string {
	n := strings.Split(rt.Path, "/")
	n = append(n, strings.ToLower(rt.Method))
	for index, value := range n {
		if regSplat.MatchString(value) {
			n[index] = "{s}"
		}
		if regParam.MatchString(value) {
			n[index] = "{p}"
		}
	}
	return strings.Join(n, `\`)
}

var regParam = regexp.MustCompile(`:[^/#?()\.\\]+|\(\?P<[a-zA-Z0-9]+>.*\)`)
var regSplat = regexp.MustCompile(`\*[^/#?()\.\\]+|\(\?P<[a-zA-Z0-9]+>.*\)`)

// Url returns a url for the route, provided the string parameters.
func (rt *Route) Url(params ...string) (*url.URL, error) {
	paramCount := len(params)
	i := 0
	rurl := regParam.ReplaceAllStringFunc(rt.Path, func(m string) string {
		var val string
		if i < paramCount {
			val = params[i]
		}
		i += 1
		return fmt.Sprintf(`%s`, val)
	})
	rurl = regSplat.ReplaceAllStringFunc(rurl, func(m string) string {
		splat := params[i:(len(params))]
		i += len(splat)
		return fmt.Sprintf(strings.Join(splat, "/"))
	})
	u, err := url.Parse(rurl)
	if err != nil {
		return nil, err
	}
	if i < len(params) && rt.Method == "GET" {
		providedquerystring := params[i:(len(params))]
		var querystring []string
		qsi := 0
		for qi, qs := range providedquerystring {
			if len(strings.Split(qs, "=")) != 2 {
				qs = fmt.Sprintf("value%d=%s", qi+1, qs)
			}
			querystring = append(querystring, url.QueryEscape(qs))
			qsi += 1
		}
		u.RawQuery = strings.Join(querystring, "&")
	}
	return u, nil
}
