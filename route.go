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

var (
	regParam = regexp.MustCompile(`:[^/#?()\.\\]+|\(\?P<[a-zA-Z0-9]+>.*\)`)
	regSplat = regexp.MustCompile(`\*[^/#?()\.\\]+|\(\?P<[a-zA-Z0-9]+>.*\)`)
)

type (
	MakeCtxFunc func(w http.ResponseWriter, rq *http.Request, rs *engine.Result, rt *Route) Ctx

	Route struct {
		registered bool
		blueprint  *Blueprint
		static     bool
		method     string
		base       string
		path       string
		managers   []Manage
		Name       string
		MakeCtx    MakeCtxFunc
	}

	Routes map[string]*Route
)

func (app *App) Routes() Routes {
	allroutes := make(Routes)
	for _, blueprint := range app.Blueprints() {
		for _, route := range blueprint.Routes {
			if route.Name != "" {
				allroutes[route.Name] = route
			} else {
				allroutes[route.Named()] = route
			}
		}
	}
	return allroutes
}

func (app *App) existingRoute(route *Route) bool {
	for _, r := range app.Routes() {
		if route.path == r.path {
			return true
		}
	}
	return false
}

func (app *App) MergeRoutes(blueprint *Blueprint, routes Routes) {
	for _, route := range routes {
		if route.static && !app.existingRoute(route) {
			blueprint.STATIC(route.path)
		}
		if !route.static && !app.existingRoute(route) {
			blueprint.Manage(route)
		}
	}
}

func (rt *Route) App() *App {
	return rt.blueprint.app
}

func (rt *Route) rule(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
	c := rt.MakeCtx(rw, rq, rs, rt)
	c.Run()
	c.Cancel()
}

func NewRoute(method string, path string, static bool, managers []Manage) *Route {
	rt := &Route{method: method, static: static, managers: managers}
	if static {
		if fp := strings.Split(path, "/"); fp[len(fp)-1] != "*filepath" {
			rt.base = filepath.ToSlash(filepath.Join(path, "/*filepath"))
		} else {
			rt.base = path
		}
	} else {
		rt.base = path
	}
	return rt
}

func (rt *Route) Named() string {
	name := strings.Split(rt.path, "/")
	name = append(name, strings.ToLower(rt.method))
	for index, value := range name {
		if regSplat.MatchString(value) {
			name[index] = "s"
		}
		if regParam.MatchString(value) {
			name[index] = "p"
		}
	}
	return strings.Join(name, `\`)
}

func (rt *Route) Url(params ...string) (*url.URL, error) {
	paramCount := len(params)
	i := 0
	rurl := regParam.ReplaceAllStringFunc(rt.path, func(m string) string {
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
	if i < len(params) && rt.method == "GET" {
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
