package app

import (
	"mime/multipart"

	ce "github.com/thrisp/flotilla/app/extensions/cookie"
	re "github.com/thrisp/flotilla/app/extensions/response"
	se "github.com/thrisp/flotilla/app/extensions/session"
	"github.com/thrisp/flotilla/extension"
	"github.com/thrisp/flotilla/route"
	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/store"
	"github.com/thrisp/flotilla/xrr"
)

func mkFunction(k string, v interface{}) extension.Function {
	return extension.NewFunction(k, v)
}

type RequestFiles map[string][]*multipart.FileHeader

func Files(s state.State) RequestFiles {
	f, err := s.Call("files")
	if err == nil {
		if files, ok := f.(RequestFiles); ok {
			return files
		}
	}
	return nil
}

func files(s state.State) RequestFiles {
	rq := s.Request()
	if rq.MultipartForm.File != nil {
		return rq.MultipartForm.File
	}
	return nil
}

func modeIsFunc(a *App) func(s state.State, is string) bool {
	return func(s state.State, is string) bool {
		return a.GetMode(is)
	}
}

func statusFunc(a *App) func(state.State, int) error {
	return func(s state.State, code int) error {
		st := a.GetStatus(code)
		s.Rerun(st.Managers()...)
		return nil
	}
}

func storeQueryFunc(a *App) func(state.State) store.Store {
	return func(s state.State) store.Store {
		return a
	}
}

// Provide a State, Stored returns a store.Store instance.
func Stored(s state.State) store.Store {
	if st, err := s.Call("store"); err == nil {
		return st.(store.Store)
	}
	return nil
}

// Provided a State, and a key string, StoreString queries the associated Store
// returning the value corresponding to the key, or a nil string.
func StoredString(s state.State, key string) string {
	if st := Stored(s); st != nil {
		return st.String(key)
	}
	return ""
}

func routesMapFunc(a *App) func() map[string]*route.Route {
	return func() map[string]*route.Route {
		ret := make(map[string]*route.Route)
		for _, b := range a.ListBlueprints() {
			m := b.Map()
			for k, v := range m {
				ret[k] = v
			}
		}
		return ret
	}
}

var noUrl = xrr.NewXrror("Unable to get url for route %s with params %s.").Out

func urlForFunc(a *App) func(state.State, string, bool, []string) (string, error) {
	routeFn := routesMapFunc(a)
	return func(s state.State, routeName string, external bool, params []string) (string, error) {
		routes := routeFn()
		if rt, ok := routes[routeName]; ok {
			routeUrl, _ := rt.Url(params...)
			if routeUrl != nil {
				if external {
					routeUrl.Host = s.Request().Host
				}
				return routeUrl.String(), nil
			}
		}
		return "", noUrl(routeName, params)
	}
}

func stateExtension(a *App) extension.Extension {
	stateFns := []extension.Function{
		mkFunction("files", files),
		mkFunction("mode_is", modeIsFunc(a)),
		mkFunction("status", statusFunc(a)),
		mkFunction("store", storeQueryFunc(a)),
		mkFunction("stored_string", StoredString),
		mkFunction("url_for", urlForFunc(a)),
	}

	return extension.New("State_Extension", stateFns...)
}

var extensions = []extension.Extension{
	ce.Extension,
	re.Extension,
	se.Extension,
}

// Provided an App instance, BuiltInExtension returns a default
// extension.Extension that includes extensions for cookies, responses, and
// sessions, as well as several assorted App & State dependent extension
// functions.
func BuiltInExtension(a *App) extension.Extension {
	ext := extension.New("BuiltIn_Extension")
	ext.Extend(stateExtension(a))
	ext.Extend(extensions...)
	return ext
}
