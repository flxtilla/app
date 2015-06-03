package flotilla

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/thrisp/flotilla/session"
)

func TestExtension(t *testing.T) {}

func TestCookieExtension(t *testing.T) {
	a := testApp(t, "testCookieExtension", nil, nil)

	cookiem := func(c Ctx) {
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

	a.GET("/cookies_test", cookiem)
	a.Configure(a.Configuration...)

	req, _ := http.NewRequest("GET", "/cookies_test", nil)

	req.AddCookie(&http.Cookie{Name: "GetCookie1", Value: "cookie value"})

	v := securevalue("Flotilla;Secret;Key;1", "cookie value")
	req.AddCookie(&http.Cookie{Name: "GetCookie2", Value: v})

	res := httptest.NewRecorder()

	a.ServeHTTP(res, req)
}

func TestResponseExtension(t *testing.T) {
	a := testApp(t, "testResponseExtension", nil, nil)
	rm1 := func(c Ctx) { c.Call("abort", 406) }
	rm2 := func(c Ctx) {
		c.Call("headerwrite", 406)
		c.Call("headernow")
	}
	rm3 := func(c Ctx) {
		c.Call("headermodify", "add", []string{"From", "testing"})
		c.Call("writetoresponse", "written from testing")
		c.Call("headernow")
		written, _ := c.Call("iswritten")
		if !written.(bool) {
			t.Errorf("iswritten returned %t, but should be true", written)
		}
	}
	rm4 := func(c Ctx) {
		c.Call("serveplain", 200, "serveplain from testing")
	}
	rm5 := func(c Ctx) {
		c.Call("redirect", 303, "/redirectedto")
	}
	a.GET("/rt1/", rm1) //abort
	a.GET("/rt2/", rm2) //headerwrite & headernow
	a.GET("/rt3/", rm3) //headermodify & writetoresponse
	a.GET("/rt4/", rm4) //serveplain
	//servefile tested with static tests
	a.GET("/rt5/", rm5) //redirect
	a.Configure()
	p := NewPerformer(t, a, 406, "GET", "/rt1/")
	performFor(p)
	p = NewPerformer(t, a, 406, "GET", "/rt2/")
	performFor(p)
	p = NewPerformer(t, a, 200, "GET", "/rt3/")
	performFor(p)
	if p.response.HeaderMap["From"][0] != "testing" {
		t.Errorf("Header was not modified correctly, expected From = testing, but header only contained %s", p.response.HeaderMap)
	}
	if !strings.Contains(p.response.Body.String(), "written from testing") {
		t.Errorf(`Expected "written from testing" in repsonse body:\n   %s`, p.response.Body)
	}
	p = NewPerformer(t, a, 200, "GET", "/rt4/")
	performFor(p)
	if !strings.Contains(p.response.Body.String(), "serveplain from testing") {
		t.Errorf(`Expected "serveplain from testing" in repsonse body:\n   %s`, p.response.Body)
	}
	p = NewPerformer(t, a, 303, "GET", "/rt5/")
	performFor(p)
	rd := p.response.HeaderMap["Location"][0]
	if rd != "/redirectedto" {
		t.Errorf(`Expected redirect to location "/redirectedto", but received %s`, rd)
	}
}

func TestSessionExtension(t *testing.T) {
	a := testApp(t, "testSessionExtension", nil, nil)
	sm := func(c Ctx) {
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
	a.GET("/st1/", sm)
	a.Configure()
	p := NewPerformer(t, a, 200, "GET", "/st1/")
	performFor(p)
}

func TestFlashExtension(t *testing.T) {
	a := testApp(t, "testFlash", nil, nil)
	fm := func(c Ctx) {
		//c.Call("flash", "flashed", "flashed test message")
		//fd, _ := c.Call("flashed")
		//if fd, ok := fd.(map[string][]string); ok {
		//	rfd := fd["flashed"]
		//	if len(rfd) == 1 {
		//		if rfd[0] != "flashed test message" {
		//			t.Errorf(`extension: flashed returned %s, not "flashed test message"`, rfd)
		//		}
		//	} else {
		//		t.Errorf(`extension: flashed returned len %d, not 1`, len(rfd))
		//	}
		//} else {
		//	t.Errorf("%+v was not type map[string][]string", fd)
		//}
		//c.Call("flash", "cat1", "category one")
		//c.Call("flash", "cat2", "category two")
		//fs, _ := c.Call("flashes", []string{"cat1", "cat2", "cat3"})
		//if fs, ok := fs.(map[string][]string); ok {
		//	one := fs["cat1"][0]
		//	two := fs["cat2"][0]
		//	_, three := fs["cat3"]
		//	if one != "category one" || two != "category two" || three {
		//		t.Errorf("extension: flashes expected map[cat1:[category one] cat2:[category two]], received %+v", fs)
		//	}
		//} else {
		//	t.Errorf("%+v was not type map[string][]string", fs)
		//}
		//c.Call("flash", "test", "test flash")
	}
	a.GET("/ft1/", fm)
	p := NewPerformer(t, a, 200, "GET", "/ft1/")
	performFor(p)
}

func TestCtxExtension(t *testing.T) {
	a := testApp(t, "testCtxExtensions", nil, nil)
	type td struct {
		a string
		b int
	}
	cm1 := func(c Ctx) {
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
	a.GET("/ct1/", cm1)
	p := NewPerformer(t, a, 200, "GET", "/ct1/")
	performFor(p)

	cm2 := func(c Ctx) {
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
	a.GET("/ct2/", cm2)
	p = NewPerformer(t, a, 200, "GET", "/ct2/")
	performFor(p)

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
	a := testApp(
		t,
		"testAddedExtension",
		testConf(
			Extensions(ExtensionForTest),
		),
		nil,
	)

	var ret1, ret2 string

	extfunc := func(c Ctx) {
		r1, _ := c.Call("ext0")
		ret1 = r1.(string)
		r2, _ := c.Call("ext1")
		ret2 = r2.(string)
	}

	a.GET("/extension", extfunc)

	a.Configure(a.Configuration...)

	p := NewPerformer(t, a, 200, "GET", "/extension")

	performFor(p)

	if ret1 != "RETURNED FROM TESTEXTENSIONZERO" {
		t.Errorf(`Test extension 0 incorrectly return %s`, ret1)
	}

	if ret2 != "RETURNED FROM TESTEXTENSIONONE" {
		t.Errorf(`Test extension 1 incorrectly return %s`, ret2)
	}
}
