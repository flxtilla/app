package static

/*
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
		WithAssets(TestAsset),
	)
	a.STATIC("/resources/static/css/*filepath")
	a.StaticDirs(teststaticdirectory())
	ZeroExpectationPerformer(t, a, 200, "GET", "/static/css/static.css").Perform()
	ZeroExpectationPerformer(t, a, 200, "GET", "/static/css/css/css/css_asset.css").Perform()
	ZeroExpectationPerformer(t, a, 404, "GET", "/static/css/no.css").Perform()
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
		UseStaticor(ss),
	)
	a.STATIC("/staticor/")
	p := ZeroExpectationPerformer(t, a, 200, "GET", "/staticor/")
	p.Perform()
	b := p.response.Body.String()
	if b != "from external staticor" {
		t.Errorf(`Test external staticor did not return "from external staticor", returned %s`, b)
	}
}
*/
