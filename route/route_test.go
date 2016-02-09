package route_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thrisp/flotilla/app"
	"github.com/thrisp/flotilla/route"
	"github.com/thrisp/flotilla/state"
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

func one(s state.State) {}

func two(s state.State) {}

func three(s state.State) {}

func TestRoute(t *testing.T) {
	r1 := route.New(route.DefaultRouteConf("GET", "/one/:route", []state.Manage{one, two}))

	r2 := route.New(route.DefaultRouteConf("GET", "/two/:route", []state.Manage{one, two}))
	r2.Configure(
		func(r *route.Route) error {
			r.Rename("NamedRoute")
			return nil
		},
	)

	r3 := route.New(route.StaticRouteConf("GET", "/stc/*filepath", []state.Manage{three}))

	r4 := route.New(route.DefaultRouteConf("POST", "/random/route/with/:param", []state.Manage{one, two}))

	a := AppForTest(t, "testroute")

	a.Manage(r1)
	a.Manage(r2)
	a.Manage(r3)
	a.Manage(r4)

	name1, name2, name3, name4 := r1.Name(), r2.Name(), r3.Name(), r4.Name()

	names := strings.Join([]string{name1, name2, name3, name4}, ",")

	keys := []string{name1, name2, name3, name4}

	expected := strings.Join([]string{`\one\{p}\get`, `NamedRoute`, `\stc\{s}\get`, `\random\route\with\{p}\post`}, ",")

	if bytes.Compare([]byte(names), []byte(expected)) != 0 {
		t.Errorf(`Route names were [%s], but should be ["\one\{p}\get", "NamedRoute", "\stc\{s}\get", "\random\route\with\{p}\post"]`, names)
	}

	rts := a.Blueprints.Map()

	for _, key := range keys {
		if _, ok := rts[key]; !ok {
			t.Errorf(`%s was not found in App.Routes()`, key)
		}
	}

	url1, _ := r1.Url("parameter_one/")
	url2, _ := r2.Url("parameter_two", "v1=another", "also=this", "with")
	url3, _ := r3.Url("static/file/path/splat")
	url4, _ := r4.Url("a_parameter")

	urls := strings.Join([]string{url1.String(), url2.String(), url3.String(), url4.String()}, ",")

	expected = "/one/parameter_one/,/two/parameter_two?v1%3Danother&also%3Dthis&value3%3Dwith,/stc/static/file/path/splat,/random/route/with/a_parameter"

	if bytes.Compare([]byte(urls), []byte(expected)) != 0 {
		t.Errorf(`Urls were [%s], but should be [/one/parameter_one/,/two/parameter_two/,/stc/static/file/path/splat,/random/route/with/a_parameter]`, urls)
	}
}

func TestRoutes(t *testing.T) {}
