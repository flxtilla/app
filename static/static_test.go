package static

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thrisp/flotilla/app"
	"github.com/thrisp/flotilla/asset"
	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/static/resources"
	"github.com/thrisp/flotilla/txst"
)

func testLocation() string {
	wd, _ := os.Getwd()
	ld, _ := filepath.Abs(wd)
	return ld
}

var TestAsset asset.AssetFS = asset.NewAssetFS(
	resources.Asset,
	resources.AssetDir,
	resources.AssetNames,
	"",
)

func AppForTest(t *testing.T, name string, conf ...app.ConfigurationFn) *app.App {
	conf = append(conf, app.Mode("Testing", true))
	a := app.New(name, conf...)
	err := a.Configure()
	if err != nil {
		t.Errorf("Error in app configuration: %s", err.Error())
	}
	return a
}

func TestStatic(t *testing.T) {
	a := AppForTest(
		t,
		"testStatic",
		app.Assets(TestAsset),
	)

	a.STATIC("/resources/static/*filepath", "resources", a.Environment)

	txst.ZeroExpectationPerformer(t, a, 200, "GET", "/resources/static/css/static.css").Perform()

	exp, _ := txst.NewExpectation(
		200,
		"GET",
		"/static/css/static.css",
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			b := r.Body.String()
			if !strings.Contains(b, "test css file") {
				t.Error(`Test css file did not return "test css file"`)
			}
		},
	)
	exp.SetPreRegister(true)
	txst.SimplePerformer(t, a, exp).Perform()

	txst.ZeroExpectationPerformer(t, a, 200, "GET", "/static/css/css/css/css_asset.css").Perform()
	txst.ZeroExpectationPerformer(t, a, 404, "GET", "/static/css/no.css").Perform()
}

type testStatic struct{}

func (ts *testStatic) StaticDirs(d ...string) []string {
	return []string{""}
}

func (ts *testStatic) Exists(s state.State, str string) bool {
	return true
}

func (ts *testStatic) StaticManage(s state.State) {
	s.Call("write_to_response", "from external staticr")
}

func TestCustomStaticor(t *testing.T) {
	ss := &testStatic{}
	a := AppForTest(
		t,
		"testExternalStaticor",
	)
	a.SwapStaticr(ss)
	exp, _ := txst.NewExpectation(
		200,
		"GET",
		"/test/staticr/",
		func(t *testing.T) state.Manage {
			return a.StaticManage
		},
	)
	exp.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			b := r.Body.String()
			if b != "from external staticr" {
				t.Error(`Test external staticr did not return "from external staticr"`)
			}

		},
	)
	txst.SimplePerformer(t, a, exp).Perform()
}
