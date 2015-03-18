package flotilla

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thrisp/flotilla/session"
)

func TestExtension(t *testing.T) {}

func TestCookieExtension(t *testing.T) {
	a := New("testCookieExtension")

	cookiem := func(c Ctx) {
		var err error

		_, err = c.Call("cookie", false, "set-cookie", "cookie value", []interface{}{1000000, "/path", "domain.com", true, true})
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
	a := New("testResponseExtension")
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
	a := New("testSessionExtension")
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
	// a := New("testFlashExtension")
}

func TestCtxExtension(t *testing.T) {
	// a := New("testCtxExtension")
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
	a := New("testAddedExtension", Extensions(ExtensionForTest))

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
