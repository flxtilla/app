package status

import (
	"bytes"
	"fmt"

	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/xrr"
)

const statusText = `%d %s`

const panicBlock = `<h1>%s</h1>
<pre style="font-weight: bold;">%s</pre>
`
const panicHtml = `<html>
<head><title>Internal Server Error</title>
<style type="text/css">
html, body {
font-family: "Roboto", sans-serif;
color: #333333;
margin: 0px;
}
h1 {
color: #2b3848;
background-color: #ffffff;
padding: 20px;
border-bottom: 1px dashed #2b3848;
}
pre {
font-size: 1.1em;
margin: 20px;
padding: 20px;
border: 2px solid #2b3848;
background-color: #ffffff;
}
pre p:nth-child(odd){margin:0;}
pre p:nth-child(even){background-color: rgba(216,216,216,0.25); margin: 0;}
</style>
</head>
<body>
%s
</body>
</html>
`

type Status interface {
	Code() int
	Managers() []state.Manage
}

type status struct {
	code     int
	managers []state.Manage
}

func newStatus(code int, m ...state.Manage) *status {
	st := &status{code: code}
	st.managers = []state.Manage{st.first}
	st.managers = append(st.managers, m...)
	st.managers = append(st.managers, st.panics, st.last)
	return st
}

func (s *status) Code() int {
	return s.code
}

func (s *status) Managers() []state.Manage {
	return s.managers
}

func (st status) first(s state.State) {
	s.Call("header_write", st.code)
}

func panicSignal(s state.State) {
	for _, _ = range panics(s) {
		//sig := fmt.Sprintf("encountered an internal error: %s\n-----\n%s\n-----\n", p.Error(), p.Meta)
		//s.Call("panic_signal", sig)
	}
}

func panicServe(s state.State, b bytes.Buffer) {
	servePanic := fmt.Sprintf(panicHtml, b.String())
	_, _ = s.Call("header_modify", "set", []string{"Content-Type", "text/html"})
	_, _ = s.Call("write_to_response", servePanic)
}

func panics(s state.State) xrr.Xrrors {
	return s.Errors().ByType(xrr.ErrorTypePanic)
}

// Panics returns any panic type error messages attached to the Ctx.
//func Panics(s state.State) xrr.Xrrors {
//	panics, _ := s.Call("panics")
//	return panics.(xrr.Xrrors)
//}

func panicToBuffer(s state.State) bytes.Buffer {
	var auffer bytes.Buffer
	//	for _, p := range Panics(s) {
	//		reader := bufio.NewReader(bytes.NewReader([]byte(fmt.Sprintf("%s", p.Meta))))
	//		lineno := 0
	//		var buffer bytes.Buffer
	//		var err error
	//		for err == nil {
	//			lineno++
	//			l, _, err := reader.ReadLine()
	//			if lineno%2 == 0 {
	//				buffer.WriteString(fmt.Sprintf("\n%s</p>\n", l))
	//			} else {
	//				buffer.WriteString(fmt.Sprintf("<p>%s\n", l))
	//			}
	//			if err != nil {
	//				break
	//			}
	//		}
	//		pb := fmt.Sprintf(panicBlock, p.Error(), buffer.String())
	//		auffer.WriteString(pb)
	//	}
	return auffer
}

func isWritten(s state.State) bool {
	return s.RWriter().Written()
}

func (st status) panics(s state.State) {
	if st.code == 500 && !isWritten(s) {
		//if !ModeIs(s, "production") {
		//	panicServe(s, panicToBuffer(s))
		//	panicSignal(s)
		//}
	}
}

func (st status) last(s state.State) {
	//_, _ = s.Call("push", func(ps state.State) {
	//	if !IsWritten(ps) {
	//		_, _ = ps.Call("writetoresponse", fmt.Sprintf(statusText, st.code, http.StatusText(st.code)))
	//	}
	//})
}

// CustomStatusRule provides a custom engine Rule with a provided set of Manage
// to be used by the Engine in return a status.
//func CustomStatusRule(a *App, code int, m ...Manage) engine.Rule {
//	s := newStatus(code, m...)
//	a.CustomStatus(s)
//	return func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
//		s := NewState(a.Fxtension, rs)
//		s.Reset(rq, rw, s.managers)
//		s.SessionStore, _ = a.Start(s.RWriter(), s.Request())
//		s.Run()
//		s.Cancel()
//	}
//}

// HasCustomStatus returns a status and a boolean indicating existence from the
// provided App Env.
//func HasCustomStatus(a *App, code int) (*status, bool) {
//	if s, ok := a.Env.customstatus[code]; ok {
//		return s, true
//	}
//	return newStatus(code), false
//}

//func statusfunc(a *App) func(state.State, int) error {
//	return func(s state.State, code int) error {
//		st, _ := HasCustomStatus(a, code)
//		s.Rerun(st.managers...)
//		return nil
//	}
//}
