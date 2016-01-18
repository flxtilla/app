package app

import (
	//"mime/multipart"
	//"net/http"

	"mime/multipart"
	"net/http"

	"github.com/thrisp/flotilla/engine"

	"github.com/thrisp/flotilla/extension"
	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/store"
	"github.com/thrisp/flotilla/xrr"
)

func mkFunction(k string, v interface{}) extension.Function {
	return extension.NewFunction(k, v)
}

var responseFns = []extension.Function{
	mkFunction("abort", abort),
	mkFunction("header_now", headerNow),
	mkFunction("header_write", headerWrite),
	mkFunction("header_modify", headerModify),
	mkFunction("is_written", isWritten),
	mkFunction("redirect", redirect),
	mkFunction("serve_file", serveFile),
	mkFunction("serve_plain", servePlain),
	mkFunction("write_to_response", writeToResponse),
}

var ResponseFxtension extension.Fxtension = extension.New("Response_Fxtension", responseFns...)

func abort(s state.State, code int) error {
	if code >= 0 {
		w := s.RWriter()
		w.WriteHeader(code)
		w.WriteHeaderNow()
	}
	return nil
}

func headerNow(s state.State) error {
	s.RWriter().WriteHeaderNow()
	return nil
}

func headerWrite(s state.State, code int, values ...[]string) error {
	if code >= 0 {
		s.RWriter().WriteHeader(code)
	}
	headerModify(s, "set", values...)
	return nil
}

func headerModify(s state.State, action string, values ...[]string) error {
	w := s.RWriter()
	switch action {
	case "set":
		for _, v := range values {
			w.Header().Set(v[0], v[1])
		}
	default:
		for _, v := range values {
			w.Header().Add(v[0], v[1])
		}
	}
	return nil
}

func isWritten(s state.State) bool {
	return s.RWriter().Written()
}

var InvalidStatusCode = xrr.NewXrror("Cannot send a redirect with status code %d").Out

func release(s state.State) {
	w := s.RWriter()
	if !w.Written() {
		s.Out(s)
		s.SessionRelease(w)
	}
}

func redirect(s state.State, code int, location string) error {
	if code >= 300 && code <= 308 {
		s.Bounce(func(ps state.State) {
			release(ps)
			w := ps.RWriter()
			http.Redirect(w, ps.Request(), location, code)
			w.WriteHeaderNow()
		})
		return nil
	} else {
		return InvalidStatusCode(code)
	}
}

func servePlain(s state.State, code int, data string) error {
	s.Push(func(ps state.State) {
		headerWrite(ps, code, []string{"Content-Type", "text/plain"})
		ps.RWriter().Write([]byte(data))
	})
	return nil
}

func serveFile(s state.State, f http.File) error {
	fi, err := f.Stat()
	if err == nil {
		http.ServeContent(s.RWriter(), s.Request(), fi.Name(), fi.ModTime(), f)
	}
	return err
}

func writeToResponse(s state.State, data string) error {
	s.RWriter().Write([]byte(data))
	return nil
}

var sessionFns = []extension.Function{
	mkFunction("delete_session", deleteSession),
	mkFunction("get_session", getSession),
	mkFunction("session", returnSession),
	mkFunction("set_session", setSession),
}

var SessionFxtension extension.Fxtension = extension.New("Session_Fxtension", sessionFns...)

func deleteSession(s state.State, key string) error {
	return s.Delete(key)
}

func getSession(s state.State, key string) interface{} {
	return s.Get(key)
}

func SessionStore(s state.State) session.SessionStore {
	ss, _ := s.Call("session")
	return ss.(session.SessionStore)
}

func returnSession(s state.State) session.SessionStore {
	return s
}

func setSession(s state.State, key string, value interface{}) error {
	return s.Set(key, value)
}

//func statusfunc(a *App) func(*ctx, int) error {
//	return func(c *ctx, code int) error {
//		s, _ := HasCustomStatus(a, code)
//		c.rerun(s.managers...)
//		return nil
//	}
//}

func makeCtxFxtension(a *App) extension.Fxtension {
	ctxFns := []extension.Function{
		//mkFunction("env", envqueryfunc(a)),
		mkFunction("files", files),
		mkFunction("get", getData),
		mkFunction("mode_is", modeIsFunc(a)),
		//mkFunction("out", out(a)),
		//mkFunction("emit", emit(a)),
		mkFunction("panics", panics),
		//mkFunction("panic_signal", panicSignalFunc(a)),
		mkFunction("params", currentParams),
		mkFunction("param_string", paramString),
		mkFunction("push", push),
		mkFunction("render_template", renderTemplateFunc(a)),
		mkFunction("set", setData),
		//mkFunction("status", statusfunc(a)),
		//mkFunction("store", storequeryfunc(a)),
		mkFunction("urlfor", urlForFunc(a)),
	}

	return extension.New("Ctx_Fxtension", ctxFns...)
}

//func envqueryfunc(a *App) func(state.State, string) interface{} {
//	return func(s state.State, item string) interface{} {
//		switch item {
//case "store":
//	return a.Env.Store
//case "fxtensions":
//	return a.Env.Fxtension
//case "processors":
//	return a.Env.ctxprocessors
//default:
//	return nil
//		}
//	}
//}

type RequestFiles map[string][]*multipart.FileHeader

func files(s state.State) RequestFiles {
	rq := s.Request()
	if rq.MultipartForm.File != nil {
		return rq.MultipartForm.File
	}
	return nil
}

var InvalidKey = xrr.NewXrror("Key %s does not exist.").Out

func getData(s state.State, key string) (interface{}, error) {
	//item, ok := s.Data[key]
	//if ok {
	//	return item, nil
	//}
	return nil, InvalidKey(key)
}

func modeIsFunc(a *App) func(s state.State, is string) bool {
	return func(s state.State, is string) bool {
		return a.GetMode(is)
	}
}

func currentParams(s state.State) engine.Params {
	return nil //s.Params
}

func panics(s state.State) xrr.Xrrors {
	return s.Errors().ByType(xrr.ErrorTypePanic)
}

func panicSignalFunc(a *App) func(state.State, string) error {
	return func(s state.State, sig string) error {
		//a.Panic(sig)
		return nil
	}
}

func paramString(s state.State, key string) string {
	//for _, v := range c.Params {
	//	if v.Key == key {
	//		return v.Value
	//	}
	//}
	return ""
}

//func out(a *App) func(state.State, string) error {
//	return func(s state.State, msg string) error {
//		//a.Send("out", msg)
//		return nil
//	}
//}

//func emit(a *App) func(state.State, string) error {
//	return func(s state.State, msg string) error {
//		//a.Send("emit", msg)
//		return nil
//	}
//}

func push(s state.State, m state.Manage) error {
	s.Push(m)
	return nil
}

func renderTemplateFunc(a *App) func(state.State, string, interface{}) error {
	return func(s state.State, name string, data interface{}) error {
		s.Push(func(ps state.State) {
			//td := NewTemplateData(ps, data)
			//a.Render(ps.RWriter(), name, td)
		})
		return nil
	}
}

func setData(s state.State, key string, item interface{}) error {
	//if c.Data == nil {
	//	c.Data = make(map[string]interface{})
	//}
	//c.Data[key] = item
	return nil
}

func storeQueryFunc(a *App) func(state.State) store.Store {
	return func(s state.State) store.Store {
		return a
	}
}

func Stored(s state.State) store.Store {
	if st, err := s.Call("store"); err == nil {
		return st.(store.Store)
	}
	return nil
}

func StoredString(s state.State, key string) string {
	if st := Stored(s); st != nil {
		return st.String(key)
	}
	return ""
}

var NoUrl = xrr.NewXrror("Unable to get url for route %s with params %s.").Out

func urlForFunc(a *App) func(state.State, string, bool, []string) (string, error) {
	return func(s state.State, route string, external bool, params []string) (string, error) {
		//if route, ok := a.Routes()[route]; ok {
		//	routeurl, _ := route.Url(params...)
		//	if routeurl != nil {
		//		if external {
		//			routeurl.Host = s.Request().Host
		//		}
		//		return routeurl.String(), nil
		//	}
		//}
		return "", NoUrl(route, params)
	}
}

var readyFxtensions = []extension.Fxtension{
	CookieFxtension,
	//FlashFxtension,
	ResponseFxtension,
	SessionFxtension,
}

func BuiltInExtension(a *App) extension.Fxtension {
	ext := extension.New("BuiltIn_Fxtension")
	//ext.Extend(readyFxtensions...)
	//ext.Extend(makeCtxFxtension(a))
	return ext
}
