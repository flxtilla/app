package flotilla

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

	"github.com/thrisp/flotilla/engine"
	"github.com/thrisp/flotilla/xrr"
)

const (
	statusText = `%d %s`

	panicBlock = `<h1>%s</h1>
<pre style="font-weight: bold;">%s</pre>
`
	panicHtml = `<html>
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
)

type (
	status struct {
		code     int
		managers []Manage
	}
)

func (s status) first(c Ctx) {
	c.Call("headerwrite", s.code)
}

func panicsignal(c Ctx) {
	if !CurrentMode(c).Testing {
		for _, p := range Panics(c) {
			sig := fmt.Sprintf("encountered an internal error: %s\n-----\n%s\n-----\n", p.Error(), p.Meta)
			c.Call("panicsignal", sig)
		}
	}
}

func panicserve(c Ctx, b bytes.Buffer) {
	servePanic := fmt.Sprintf(panicHtml, b.String())
	_, _ = c.Call("headermodify", "set", []string{"Content-Type", "text/html"})
	_, _ = c.Call("writetoresponse", servePanic)
}

// Panics returns any panic type error messages attached to the Ctx.
func Panics(c Ctx) xrr.ErrorMsgs {
	panics, _ := c.Call("panics")
	return panics.(xrr.ErrorMsgs)
}

func panictobuffer(c Ctx) bytes.Buffer {
	var auffer bytes.Buffer
	for _, p := range Panics(c) {
		reader := bufio.NewReader(bytes.NewReader([]byte(fmt.Sprintf("%s", p.Meta))))
		lineno := 0
		var buffer bytes.Buffer
		var err error
		for err == nil {
			lineno++
			l, _, err := reader.ReadLine()
			if lineno%2 == 0 {
				buffer.WriteString(fmt.Sprintf("\n%s</p>\n", l))
			} else {
				buffer.WriteString(fmt.Sprintf("<p>%s\n", l))
			}
			if err != nil {
				break
			}
		}
		pb := fmt.Sprintf(panicBlock, p.Error(), buffer.String())
		auffer.WriteString(pb)
	}
	return auffer
}

// IsWritten returns a boolean value indicating whether the Ctx ResponseWriter
// has been written to at the point in time the function is called.
func IsWritten(c Ctx) bool {
	ret, _ := c.Call("iswritten")
	return ret.(bool)
}

func (s status) panics(c Ctx) {
	if s.code == 500 && !IsWritten(c) {
		if !CurrentMode(c).Production {
			panicserve(c, panictobuffer(c))
			panicsignal(c)
		}
	}
}

func (s status) last(c Ctx) {
	_, _ = c.Call("push", func(c Ctx) {
		if !IsWritten(c) {
			_, _ = c.Call("writetoresponse", fmt.Sprintf(statusText, s.code, http.StatusText(s.code)))
		}
	})
}

// StatusRule provides a default engine.Rule used as a status by the Engine.
func StatusRule(a *App) engine.Rule {
	return func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
		s := newStatus(rs.Code)
		c := NewCtx(a.fxtensions, rs)
		c.reset(rq, rw, s.managers)
		c.Run()
		c.Cancel()
	}
}

// CustomStatusRule provides a custom engine Rule with a provided set of Manage
// to be used by the Engine in return a status.
func CustomStatusRule(a *App, code int, m ...Manage) engine.Rule {
	s := newStatus(code, m...)
	a.CustomStatus(s)
	return func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
		c := NewCtx(a.fxtensions, rs)
		c.reset(rq, rw, s.managers)
		c.Call("start", a.SessionManager)
		c.Run()
		c.Cancel()
	}
}

func newStatus(code int, m ...Manage) *status {
	s := &status{code: code}
	s.managers = []Manage{s.first}
	s.managers = append(s.managers, m...)
	s.managers = append(s.managers, s.panics, s.last)
	return s
}

// HasCustomStatus returns a status and a boolean indicating existence from the
// provided App Env.
func HasCustomStatus(a *App, code int) (*status, bool) {
	if s, ok := a.Env.customstatus[code]; ok {
		return s, true
	}
	return newStatus(code), false
}

func statusfunc(a *App) func(*ctx, int) error {
	return func(c *ctx, code int) error {
		s, _ := HasCustomStatus(a, code)
		c.rerun(s.managers...)
		return nil
	}
}
