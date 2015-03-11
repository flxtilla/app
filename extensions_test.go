package flotilla

import "testing"

func testextensionzero(c Ctx) string {
	return "RETURNED FROM TESTEXTENSIONZERO"
}

func testextensionone(c Ctx) string {
	return "RETURNED FROM TESTEXTENSIONONE"
}

var exts map[string]interface{} = map[string]interface{}{
	"ext1": testextensionone,
}

func TestExtension(t *testing.T) {
	a := New(
		"testExtension",
		Extension("ext0", testextensionzero),
		Extensions(exts),
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
