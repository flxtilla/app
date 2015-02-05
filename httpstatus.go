package flotilla

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
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
	statusmanager struct {
		code     int
		managers []Manage
	}
)

func (sm statusmanager) first(c *Ctx) {
	c.RW.WriteHeader(sm.code)
}

func panicsignal(c *Ctx) {
	for _, p := range c.Errors.ByType(ErrorTypePanic) {
		sig := fmt.Sprintf("encountered an internal error: %s\n-----\n%s\n-----\n", p.Err, p.Meta)
		// c.App.Panic(sig)
		fmt.Printf(sig)
	}
}

func panicserve(c *Ctx, b bytes.Buffer) {
	servePanic := fmt.Sprintf(panicHtml, b.String())
	c.RW.Header().Set("Content-Type", "text/html")
	c.RW.Write([]byte(servePanic))
}

func panictobuffer(c *Ctx) bytes.Buffer {
	var auffer bytes.Buffer
	for _, p := range c.Errors.ByType(ErrorTypePanic) {
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
		pb := fmt.Sprintf(panicBlock, p.Err, buffer.String())
		auffer.WriteString(pb)
	}
	return auffer
}

func (sm statusmanager) panics(c *Ctx) {
	if sm.code == 500 && !c.RW.Written() {
		//if !c.App.Mode.Production {
		panicserve(c, panictobuffer(c))
		panicsignal(c)
		//}
	}
}

func (sm statusmanager) last(c *Ctx) {
	c.Push(func(c *Ctx) {
		if !c.RW.Written() {
			c.RW.Write([]byte(fmt.Sprintf(statusText, sm.code, http.StatusText(sm.code))))
		}
	})
}

func (sm *statusmanager) handle(c *Ctx) {
	c.ReRun(sm.managers...)
}

func statusManage(code int, ms ...Manage) Manage {
	sm := statusmanager{code: code}
	sm.managers = []Manage{sm.first}
	sm.managers = append(sm.managers, ms...)
	sm.managers = append(sm.managers, sm.panics, sm.last)
	return sm.handle
}
