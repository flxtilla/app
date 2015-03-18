package flotilla

import (
	"bytes"
	"strings"
	"testing"
)

func one(c Ctx) {}

func two(c Ctx) {}

func three(c Ctx) {}

func TestRoute(t *testing.T) {
	r1 := NewRoute("GET", "/one/:route", false, []Manage{one, two})

	r2 := NewRoute("GET", "/two/:route", false, []Manage{one, two})
	r2.Name = "NamedRoute"

	r3 := NewRoute("GET", "/stc/*filepath", true, []Manage{three})

	r4 := NewRoute("POST", "/random/route/with/:param", false, []Manage{one, two})

	a := testApp("testroute", r1, r2)

	a.Manage(r3)
	a.Manage(r4)

	name1, name2, name3, name4 := r1.Named(), r2.Name, r3.Named(), r4.Named()

	names := strings.Join([]string{name1, name2, name3, name4}, ",")

	keys := []string{name1, name2, name3, name4}

	expected := strings.Join([]string{"\\one\\p\\get", "NamedRoute", "\\stc\\s\\get", "\\random\\route\\with\\p\\post"}, ",")

	if bytes.Compare([]byte(names), []byte(expected)) != 0 {
		t.Errorf(`Route names were [%s], but should be ["\one\p\get", "NamedRoute", "\stc\s\get", "\random\route\with\p\post"]`, names)
	}

	rts := a.Routes()

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
