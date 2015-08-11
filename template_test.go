package flotilla

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http/httptest"
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

func tplfuncsconf(tf map[string]interface{}) Configuration {
	return func(a *App) error {
		a.Env.AddTplFuncs(tf)
		return nil
	}
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

func templatecontains(t *testing.T, body string, mustcontain string) {
	if !strings.Contains(body, mustcontain) {
		t.Errorf(`Test template was not rendered correctly, expecting %s but it is not present:
		%s
		`, mustcontain, body)
	}
}

func TestDefaultTemplating(t *testing.T) {
	a := testApp(
		t,
		"testDefaultTemplating",
		WithAssets(TestAsset),
		tplfuncsconf(tplfuncs),
		CtxProcessor("Hi1", tstrg),
		CtxProcessors(tstctxprc),
	)
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

	exp1, _ := NewExpectation(
		200,
		"GET",
		"/template",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				ret := make(map[string]interface{})
				ret["Title"] = "rendered from test template test.html"
				c.Call("flash", "test_flash", "TEST_FLASH_ONE")
				c.Call("flash", "test_flash", "TEST_FLASH_TWO")
				c.Call("set", "set_in_ctx", "SET_IN_CTX")
				c.Call("rendertemplate", "test.html", ret)
			}
		},
	)
	exp1.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			lookfor := []string{
				`<div>TEST TEMPLATE</div>`,
				`Hello World!: TEST`,
				`returned STRING`,
				`<div>returned HTML</div>`,
				`returned CALL`,
				`[TEST_FLASH_ONE TEST_FLASH_TWO]`,
				`/template?value1%3Dadditional`,
				`Unable to get url for route \does\not\exist\p\s\get with params [param /a/splat/fragment].`,
				`SET_IN_CTX`,
			}

			for _, lf := range lookfor {
				templatecontains(t, r.Body.String(), lf)
			}

		},
	)

	exp2, _ := NewExpectation(
		200,
		"GET",
		"/asset_template",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				c.Call("rendertemplate", "test_asset.html", "rendered from test template test_asset.html")
			}
		},
	)
	exp2.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			templatecontains(t, r.Body.String(), `<title>rendered from test template test_asset.html</title>`)
		},
	)

	MultiPerformer(t, a, exp1, exp2).Perform()
}

type testtemplator struct{}

func (tt *testtemplator) Render(w io.Writer, s string, i interface{}) error {
	_, err := w.Write([]byte("test templator"))
	if err != nil {
		return err
	}
	return nil
}

func (tt *testtemplator) ListTemplateDirs() []string {
	return []string{"test"}
}

func (tt *testtemplator) ListTemplates() []string {
	return []string{"test_template"}
}

func (tt *testtemplator) UpdateTemplateDirs(...string) {}

func TestTemplator(t *testing.T) {
	ttr := &testtemplator{}
	a := testApp(
		t,
		"testExternalTemplator",
		WithTemplator(ttr),
	)
	exp, _ := NewExpectation(
		200,
		"GET",
		"/templator/",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				c.Call("rendertemplate", "test.html", "test data")
			}
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			b := r.Body.String()
			if b != "test templator" {
				t.Errorf(`Test external templator rendered %s, not "test templator"`, b)
			}

		},
	)
	SimplePerformer(t, a, exp).Perform()
}
