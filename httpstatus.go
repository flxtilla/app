package flotilla

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
)

const (
	statusText = `%d %s`

	statusHtml = `<!DOCTYPE HTML>
<title>%d %s</title>
<h1>%d - %s</h1>
`
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

func defaultStatuses(trees map[string]*node) {
	sts := new(node)
	sts.addRoute("403", statusManage(403))
	sts.addRoute("404", statusManage(404))
	sts.addRoute("405", statusManage(405))
	sts.addRoute("500", panicManage)
	trees["STATUS"] = sts
}

func statusManage(code int) Manage {
	message := http.StatusText(code)
	return func(c *Ctx) {
		c.RW.WriteHeader(code)
		//if c.App.HtmlStatus {
		//	c.RW.Header().Set("Content-Type", "text/html")
		//	c.RW.Write([]byte(fmt.Sprintf(statusHtml, code, message, code, message)))
		//} else {
		c.RW.Write([]byte(fmt.Sprintf(statusText, code, message)))
		//}
	}
}

func panicManage(c *Ctx) {
	c.RW.WriteHeader(500)
	panics := c.Errors.ByType(ErrorTypePanic)
	var auffer bytes.Buffer
	for _, p := range panics {
		sig := fmt.Sprintf("encountered an internal error: %s\n-----\n%s\n-----\n", p.Err, p.Meta)
		c.App.Panic(sig)
		if !c.App.Mode.Production {
			reader := bufio.NewReader(bytes.NewReader([]byte(fmt.Sprintf("%s", p.Meta))))
			var err error
			lineno := 0
			var buffer bytes.Buffer
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
	}
	if !c.App.Mode.Production {
		servePanic := fmt.Sprintf(panicHtml, auffer.String())
		c.RW.Header().Set("Content-Type", "text/html")
		c.RW.Write([]byte(servePanic))
	} else {
		c.RW.Write([]byte(fmt.Sprintf(statusText, 500, http.StatusText(500))))
	}
}
