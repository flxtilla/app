// package engine provides an interface for flotilla package routing in addition
// to a default routing engine used by default flotilla app instances.
package engine

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/thrisp/flotilla/xrr"
)

type Rule func(http.ResponseWriter, *http.Request, *Result)

type Engine interface {
	Handle(string, string, Rule)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type conf struct {
	RedirectTrailingSlash bool
	RedirectFixedPath     bool
}

type engine struct {
	*conf
	trees      map[string]*node
	StatusRule Rule
}

func defaultConf() *conf {
	return &conf{
		RedirectTrailingSlash: true,
		RedirectFixedPath:     true,
	}
}

func DefaultEngine(status Rule) *engine {
	return &engine{
		conf:       defaultConf(),
		StatusRule: status,
	}
}

func (e *engine) Handle(method string, path string, r Rule) {
	if method != "STATUS" && path[0] != '/' {
		panic("path must begin with '/'")
	}

	if e.trees == nil {
		e.trees = make(map[string]*node)
	}

	root := e.trees[method]

	if root == nil {
		root = new(node)
		e.trees[method] = root
	}

	root.addRoute(path, r)
}

func (e *engine) lookup(method, path string) *Result {
	if root := e.trees[method]; root != nil {
		if rule, params, tsr := root.getValue(path); rule != nil {
			return NewResult(200, rule, params, tsr)
		} else if method != "CONNECT" && path != "/" {
			code := 301
			if method != "GET" {
				code = 307
			}
			if tsr && e.RedirectTrailingSlash {
				var newpath string
				if path[len(path)-1] == '/' {
					newpath = path[:len(path)-1]
				} else {
					newpath = path + "/"
				}
				return NewResult(code, func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
					rq.URL.Path = newpath
					http.Redirect(rw, rq, rq.URL.String(), code)
				}, nil, tsr)
			}
			if e.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					e.RedirectTrailingSlash,
				)
				if found {
					return NewResult(code, func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
						rq.URL.Path = string(fixedPath)
						http.Redirect(rw, rq, rq.URL.String(), code)
					}, nil, tsr)
				}
			}
		}
	}
	for method := range e.trees {
		if method == method {
			continue
		}
		handle, _, _ := e.trees[method].getValue(path)
		if handle != nil {
			return e.status(405)
		}
	}
	return e.status(404)
}

func (e *engine) status(code int) *Result {
	if root := e.trees["STATUS"]; root != nil {
		if rule, params, tsr := root.getValue(strconv.Itoa(code)); rule != nil {
			return NewResult(code, rule, params, tsr)
		}
	}
	return NewResult(code, e.defaultStatus(code), nil, false)
}

func (e *engine) defaultStatus(code int) Rule {
	if e.StatusRule == nil {
		return func(rw http.ResponseWriter, rq *http.Request, rs *Result) {
			rw.WriteHeader(rs.Code)
			rw.Write([]byte(fmt.Sprintf("%d %s", rs.Code, http.StatusText(rs.Code))))
		}
	}
	return e.StatusRule
}

func (e *engine) rcvr(rw http.ResponseWriter, rq *http.Request) {
	if rcv := recover(); rcv != nil {
		s := e.status(500)
		s.Frror("%s", xrr.ErrorTypePanic, xrr.Stack(3), rcv)
		s.Rule(rw, rq, s)
	}
}

func (e *engine) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	defer e.rcvr(rw, rq)
	rslt := e.lookup(rq.Method, rq.URL.Path)
	rslt.Rule(rw, rq, rslt)
}
