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
	for _, p := range Panics(c) {
		sig := fmt.Sprintf("encountered an internal error: %s\n-----\n%s\n-----\n", p.Err, p.Meta)
		c.Call("panicsignal", sig)
	}
}

func panicserve(c Ctx, b bytes.Buffer) {
	servePanic := fmt.Sprintf(panicHtml, b.String())
	_, _ = c.Call("headermodify", "set", []string{"Content-Type", "text/html"})
	_, _ = c.Call("writetoresponse", []byte(servePanic))
}

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
			_, _ = c.Call("writetoresponse", []byte(fmt.Sprintf(statusText, s.code, http.StatusText(s.code))))
		}
	})
}

func StatusRule(a *App) engine.Rule {
	return func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
		s := newStatus(rs.Code)
		c := NewCtx(a.extensions, rs)
		c.reset(rq, rw, s.managers)
		c.Run()
		c.Cancel()
	}
}

func CustomStatusRule(a *App, code int, m ...Manage) engine.Rule {
	s := newStatus(code, m...)
	a.CustomStatus(s)
	return func(rw http.ResponseWriter, rq *http.Request, rs *engine.Result) {
		c := NewCtx(a.extensions, rs)
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

func HasCustomStatus(a *App, code int) (*status, bool) {
	s, ok := a.Env.customstatus[code]
	if !ok {
		return newStatus(code), false
	}
	return s, true
}

func makehttpstatus(a *App) func(*ctx, int) error {
	return func(c *ctx, code int) error {
		s, _ := HasCustomStatus(a, code)
		c.rerun(s.managers...)
		return nil
	}
}
