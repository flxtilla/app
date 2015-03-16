package flotilla

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testlocation() string {
	wd, _ := os.Getwd()
	ld, _ := filepath.Abs(wd)
	return ld
}

func testtemplatedirectory() string {
	return filepath.Join(testlocation(), "resources", "templates")
}

func trimtemplates(t []string) []string {
	var templates []string
	for _, f := range t {
		templates = append(templates, filepath.Base(f))
	}
	return templates
}

func stringtemplates(t []string) string {
	templates := trimtemplates(t)
	return strings.Join(templates, ",")
}

var tplfuncs map[string]interface{} = map[string]interface{}{
	"Hello": func(s string) string { return fmt.Sprintf("Hello World!: %s", s) },
}

func tstrg(c Ctx) string {
	return fmt.Sprintf("returned STRING")
}

func thtml(c Ctx) template.HTML {
	return "<div>returned HTML</div>"
}

func tcall(c Ctx) string {
	return "returned CALL"
}

var tstctxprc map[string]interface{} = map[string]interface{}{
	"Hi1": tstrg,
	"Hi2": thtml,
	"Hi3": tcall,
}

func tt(c Ctx) {
	ret := make(map[string]interface{})
	ret["Title"] = "rendered from test template test.html"
	c.Call("flash", "test_flash", "TEST_FLASH_ONE")
	c.Call("flash", "test_flash", "TEST_FLASH_TWO")
	c.Call("set", "set_in_ctx", "SET_IN_CTX")
	c.Call("rendertemplate", "test.html", ret)
}

func at(c Ctx) {
	c.Call("rendertemplate", "test_asset.html", "rendered from test template test_asset.html")
}

func templatecontains(t *testing.T, body string, mustcontain string) {
	if !strings.Contains(body, mustcontain) {
		t.Errorf(`Test template was not rendered correctly, expecting %s but it is not present:
		%s
		`, mustcontain, body)
	}
}

func TestDefaultTemplating(t *testing.T) {
	a := New("template",
		Mode("testing", true),
		WithAssets(TestAsset),
		CtxProcessors(tstctxprc),
	)

	a.Env.AddTplFuncs(tplfuncs)

	a.GET("/template", tt)
	a.GET("/asset_template", at)

	a.Configure()

	a.Env.TemplateDirs(testtemplatedirectory())

	d1 := stringtemplates(a.Env.TemplateDirs())
	d2 := stringtemplates(a.Env.Templator.ListTemplateDirs())
	if bytes.Compare([]byte(d1), []byte(d2)) != 0 {
		t.Errorf(`App Env template directories and Templator template directories differ where they should be equal.
		Env template directories:
			%s
		Templator template directories:
			%s`, d1, d2)
	}

	var expected []string = []string{"layout.html", "test.html", "layout_asset.html", "test_asset.html"}
	var templates []string = trimtemplates(a.Env.Templator.ListTemplates())

	for _, ex := range expected {
		if !existsIn(ex, templates) {
			t.Errorf(`Existing templates do not contain %s`, ex)
		}
	}

	p := NewPerformer(t, a, 200, "GET", "/template")

	performFor(p)

	lookfor := []string{
		`<div>TEST TEMPLATE</div>`,
		`Hello World!: TEST`,
		`returned STRING`,
		`<div>returned HTML</div>`,
		`returned CALL`,
		`[TEST_FLASH_ONE TEST_FLASH_TWO]`,
		`/template?value1%3Dadditional`,
		`unable to get url for route \does\not\exist\p\s\get with params [param /a/splat/fragment]`,
		`SET_IN_CTX`,
	}

	for _, lf := range lookfor {
		templatecontains(t, p.response.Body.String(), lf)
	}

	p = NewPerformer(t, a, 200, "GET", "/asset_template")

	performFor(p)

	templatecontains(t, p.response.Body.String(), `<title>rendered from test template test_asset.html</title>`)
}
