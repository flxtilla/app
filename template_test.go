package flotilla

import (
	"bytes"
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

func stringtemplates(t []string) string {
	var templates []string
	for _, f := range t {
		templates = append(templates, filepath.Base(f))
	}
	return strings.Join(templates, ",")
}

func tt(c Ctx) {
	c.Call("rendertemplate", "test.html", "rendered from test template test.html")
}

func at(c Ctx) {
	c.Call("rendertemplate", "test_asset.html", "rendered from test template test_asset.html")
}

func TestDefaultTemplating(t *testing.T) {
	a := New("template", Mode("testing", true), WithAssets(TestAssets))

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

	var expected string = "layout.html,test.html,layout_asset.html,test_asset.html"
	var templates string = stringtemplates(a.Env.Templator.ListTemplates())

	if bytes.Compare([]byte(expected), []byte(templates)) != 0 {
		t.Errorf(`Template listing should be %s, but was %s`, expected, templates)
	}

	p := NewPerformer(t, a, 200, "GET", "/template")

	performFor(p)

	if !strings.Contains(p.response.Body.String(), `<div>TEST TEMPLATE</div>`) {
		t.Errorf(`Test template was not rendered correctly:
		%s
		`, p.response.Body.String())
	}

	p = NewPerformer(t, a, 200, "GET", "/asset_template")

	performFor(p)

	if !strings.Contains(p.response.Body.String(), `<title>rendered from test template test_asset.html</title>`) {
		t.Errorf(`Test asset template was not rendered correctly:
		%s
		`, p.response.Body.String())
	}
}
