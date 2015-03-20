package flotilla

import (
	"path/filepath"
	"testing"
)

func teststaticdirectory() string {
	return filepath.Join(testlocation(), "resources", "static")
}

func TestStatic(t *testing.T) {
	a := testApp(
		t,
		"testStatic",
		testConf(
			WithAssets(TestAsset),
		),
		nil,
	)
	a.STATIC("/resources/static/css/*filepath")
	a.StaticDirs(teststaticdirectory())
	a.Configure()
	p := NewPerformer(t, a, 200, "GET", "/static/css/static.css")
	performFor(p)
	p = NewPerformer(t, a, 200, "GET", "/static/css/css/css_asset.css")
	performFor(p)
	p = NewPerformer(t, a, 404, "GET", "/static/css/no.css")
	performFor(p)
}

type teststaticor struct{}

func (ts *teststaticor) StaticDirs(d ...string) []string {
	return []string{""}
}

func (ts *teststaticor) Exists(c Ctx, s string) bool {
	return true
}

func (ts *teststaticor) Manage(c Ctx) {
	c.Call("writetoresponse", "from external staticor")
}

func TestStaticor(t *testing.T) {
	ss := &teststaticor{}
	a := testApp(
		t,
		"testExternalStaticor",
		testConf(
			UseStaticor(ss),
		),
		nil,
	)
	a.STATIC("/staticor/")
	a.Configure()
	p := NewPerformer(t, a, 200, "GET", "/staticor/")
	performFor(p)
	b := p.response.Body.String()
	if b != "from external staticor" {
		t.Errorf(`Test external staticor did not return "from external staticor", returned %s`, b)
	}
}
