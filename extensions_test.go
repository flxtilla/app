package flotilla

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/thrisp/flotilla/session"
)

func TestCookieExtension(t *testing.T) {
	cookieTester := func(t *testing.T) Manage {
		return func(c Ctx) {
			var err error

			_, err = c.Call("cookie", false, "set-cookie", "cookie value", []interface{}{1000000, "/path", "domain.com", true, true})
			_, err = c.Call("cookie", false, "set-cookie", "cookie value", []interface{}{-1, "/path", "domain.com", true, true})
			if err != nil {
				t.Errorf("Cookie was not properly set: %s", err.Error())
			}

			_, err = c.Call("securecookie", "securecookie", "secure cookie value", nil)
			if err != nil {
				t.Errorf("Secure cookie was not properly set: %s", err.Error())
			}

			ckes, _ := c.Call("cookies")
			cone := ckes.(map[string]*http.Cookie)["GetCookie1"]
			if cone.Value != "cookie value" {
				t.Errorf(`Expected cookie value "cookie value" from cookies, but received %s`, cone)
			}

			rckes, _ := c.Call("readcookies")
			ctwo := rckes.(map[string]string)["GetCookie2"]
			if ctwo != "cookie value" {
				t.Errorf(`Expected cookie value "cookie value" from readcookies, but received %s`, ctwo)
			}
		}
	}

	exp, _ := NewExpectation(
		200,
		"GET",
		"/cookies_test",
		cookieTester,
	)

	exp.Request().AddCookie(&http.Cookie{Name: "GetCookie1", Value: "cookie value"})
	v := securevalue("Flotilla;Secret;Key;1", "cookie value")
	exp.Request().AddCookie(&http.Cookie{Name: "GetCookie2", Value: v})

	app := testApp(t, "testCookieExtension")

	SimplePerformer(t, app, exp).Perform()
}

func TestResponseExtension(t *testing.T) {
	app := testApp(t, "testResponseExtension")
	exp1, _ := NewExpectation(
		406,
		"GET",
		"/rt1/",
		func(t *testing.T) Manage {
			return func(c Ctx) { c.Call("abort", 406) }
		},
	)
	exp2, _ := NewExpectation(
		406,
		"GET",
		"/rt2/",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				c.Call("headerwrite", 406)
				c.Call("headernow")
			}
		},
	)
	exp3, _ := NewExpectation(
		200,
		"GET",
		"/rt3/",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				c.Call("headermodify", "add", []string{"From", "testing"})
				c.Call("writetoresponse", "written from testing")
				c.Call("headernow")
				written, _ := c.Call("iswritten")
				if !written.(bool) {
					t.Errorf("iswritten returned %t, but should be true", written)
				}
			}
		},
	)
	exp3.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			if r.HeaderMap["From"][0] != "testing" {
				t.Errorf("Header was not modified correctly, expected From = testing, but header only contained %s", r.HeaderMap)
			}
			if !strings.Contains(r.Body.String(), "written from testing") {
				t.Errorf(`Expected "written from testing" in repsonse body:\n   %s`, r.Body)
			}
		},
	)
	exp4, _ := NewExpectation(
		200,
		"GET",
		"/rt4/",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				c.Call("serveplain", 200, "serveplain from testing")
			}
		},
	)
	exp4.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			if !strings.Contains(r.Body.String(), "serveplain from testing") {
				t.Errorf(`Expected "serveplain from testing" in repsonse body:\n   %s`, r.Body)
			}
		},
	)
	exp5, _ := NewExpectation(
		303,
		"GET",
		"/rt5/",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				c.Call("redirect", 303, "/redirectedto")
			}
		},
	)
	exp5.SetPost(
		func(t *testing.T, r *httptest.ResponseRecorder) {
			rd := r.HeaderMap["Location"][0]
			if rd != "/redirectedto" {
				t.Errorf(`Expected redirect to location "/redirectedto", but received %s`, rd)
			}
		},
	)
	MultiPerformer(t, app, exp1, exp2, exp3, exp4, exp5).Perform()
}

func TestSessionExtension(t *testing.T) {
	app := testApp(t, "testSessionExtension")
	exp, _ := NewExpectation(
		200,
		"GET",
		"/st1",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				s := Session(c)
				if _, ok := s.(session.SessionStore); !ok {
					t.Errorf("Session returned from session fxtension was not type session.SessionStore")
				}
				c.Call("setsession", "test", "value")
				v1, _ := c.Call("getsession", "test")
				v := v1.(string)
				if v != "value" {
					t.Errorf(`session key returned %s, not "value"`, v)
				}
				c.Call("deletesession", "test")
				v2, _ := c.Call("getsession", "test")
				if v2 != nil {
					t.Errorf(`deleted session item with key "test" returned %s, but should be nil`, v2)
				}
			}
		},
	)
	SimplePerformer(t, app, exp).Perform()
}

func TestFlashExtension(t *testing.T) {
	a := testApp(t, "testFlash")
	exp1, _ := NewExpectation(
		200,
		"GET",
		"/ft1/",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				c.Call("flash", "testing", "test flash message")
			}
		},
	)
	exp2, _ := NewExpectation(
		200,
		"GET",
		"/ft2/",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				fl := Flshr(c)
				v := fl.Write("testing")
				if !strings.Contains(v[0], "test flash message") {
					t.Errorf(`flash messaging expected "test flash message" as first message, but it was %s`, v[0])
				}
				fl.WriteAll()
				c.Call("flash", "testing_two", "second test flash message")
			}
		},
	)
	exp3, _ := NewExpectation(
		200,
		"GET",
		"/ft3/",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				fl := Flshr(c)
				nv := fl.Write("testing")
				if nv != nil {
					t.Errorf(`flasher wrote %s for category "testing", but was expecting a nil value`, nv)
				}
				v := fl.WriteAll()
				expected := v["testing_two"][0]
				if !strings.Contains(expected, "second test flash message") {
					t.Errorf(`flash messaging expected "second test flash message" as first message for "testing_two", but it was %s`, expected)
				}
				v2 := fl.Write("testing_two")
				if v2 != nil {
					t.Errorf(`flasher wrote %s for category "testing_two", but was expecting a nil value`, v2)
				}
			}
		},
	)

	SessionPerformer(t, a, exp1, exp2, exp3).Perform()
}

func TestCtxExtension(t *testing.T) {
	app := testApp(t, "testCtxExtensions")
	type td struct {
		a string
		b int
	}
	exp1, _ := NewExpectation(
		200,
		"GET",
		"/ct1",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				c.Call("set", "test", &td{a: "data", b: 1})
				d, _ := c.Call("get", "test")
				if d, ok := d.(*td); ok {
					if d.a != "data" || d.b != 1 {
						t.Errorf("extensions: getdata & setdata -- Test data was not properly set or gotten from ctx")
					}
				}
				e, _ := c.Call("get", "none")
				if e != nil {
					t.Errorf("extensions: getdata & setdata -- expected nil, but received %+v", e)
				}
			}
		},
	)

	exp2, _ := NewExpectation(
		200,
		"GET",
		"/ct2/",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				stor, _ := c.Call("env", "store")
				if _, ok := stor.(Store); !ok {
					t.Errorf(`extensions: env("store") expected type Store, but was %+v`, stor)
				}
				exts, _ := c.Call("env", "fxtensions")
				if _, ok := exts.(map[string]Fxtension); !ok {
					t.Errorf(`extensions: env("fxtensions") expected type map[string]Fxtension, but was %+v`, exts)
				}
				prcs, _ := c.Call("env", "processors")
				if _, ok := prcs.(map[string]reflect.Value); !ok {
					t.Errorf(`extensions: env("processors") expected type map[string]reflect.Value, but was %+v`, prcs)
				}
				none, _ := c.Call("env", "none")
				if none != nil {
					t.Errorf(`extensions: env("none") expected a nil value, but was %+v`, none)
				}
			}
		},
	)
	MultiPerformer(t, app, exp1, exp2).Perform()
}

func testextensionzero(c Ctx) string {
	return "RETURNED FROM TESTEXTENSIONZERO"
}

func testextensionone(c Ctx) string {
	return "RETURNED FROM TESTEXTENSIONONE"
}

var exts map[string]interface{} = map[string]interface{}{
	"ext0": testextensionzero,
	"ext1": testextensionone,
}

var ExtensionForTest Fxtension = MakeFxtension("testextension", exts)

func TestAddedExtension(t *testing.T) {
	app := testApp(
		t,
		"testAddedExtension",
		Extensions(ExtensionForTest),
	)

	exp, _ := NewExpectation(
		200,
		"GET",
		"/extension",
		func(t *testing.T) Manage {
			return func(c Ctx) {
				var ret1, ret2 string
				r1, _ := c.Call("ext0")
				ret1 = r1.(string)
				if ret1 != "RETURNED FROM TESTEXTENSIONZERO" {
					t.Errorf(`Test extension 0 incorrectly return %s`, ret1)
				}
				r2, _ := c.Call("ext1")
				ret2 = r2.(string)
				if ret2 != "RETURNED FROM TESTEXTENSIONONE" {
					t.Errorf(`Test extension 1 incorrectly return %s`, ret2)
				}
			}
		},
	)

	SimplePerformer(t, app, exp).Perform()
}
