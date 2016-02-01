package route

import (
	"github.com/thrisp/flotilla/xrr"
)

type Routes interface {
	GetRoute(string) (*Route, error)
	SetRoute(*Route)
	All() []*Route
	Map() map[string]*Route
}

func NewRoutes() Routes {
	return &routes{
		r: make(map[string]*Route),
	}
}

type routes struct {
	r map[string]*Route
}

func (r *routes) SetRoute(rt *Route) {
	r.r[rt.Name()] = rt
}

var NonExistent = xrr.NewXrror(`Route "%s" does not exist`).Out

func (r *routes) GetRoute(key string) (*Route, error) {
	if rt, ok := r.r[key]; ok {
		return rt, nil
	}
	return nil, NonExistent(key)
}

func (r *routes) All() []*Route {
	var ret []*Route
	for _, v := range r.r {
		ret = append(ret, v)
	}
	return ret
}

func (r *routes) Map() map[string]*Route {
	return r.r
}
