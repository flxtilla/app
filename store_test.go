package flotilla

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func testconffile() string {
	return filepath.Join(testlocation(), "resources", "flotilla.conf")
}

type defaultreturns struct {
	i   int
	i64 int64
	f   float64
	b   bool
}

func TestStore(t *testing.T) {
	storeroute := func(c Ctx) {
		_, exists := CheckStore(c, "UNREAD_VALUE")
		if exists {
			t.Errorf(`Store item value exists, but shuold not.`)
		}

		confstr, _ := CheckStore(c, "CONFSTRING")
		if confstr.Value != "ONE" {
			t.Errorf(`Store item value was not "ONE", but was %s`, confstr)
		}

		df := defaultreturns{
			confstr.Int(),
			confstr.Int64(),
			confstr.Float(),
			confstr.Bool(),
		}

		if df.i != 0 || df.i64 != -1 || df.f != 0.0 || df.b != false {
			t.Errorf(`Store item did not return correct default values: %+v`, df)
		}

		confint, _ := CheckStore(c, "CONFINT")
		if confint.Int() != 2 {
			t.Errorf(`Store item value was not "ONE", but was %s`, confint)
		}

		confint64, _ := CheckStore(c, "CONFINT64")
		if confint64.Int() != 99999 {
			t.Errorf(`Store item value was not "99999", but was %s`, confint64)
		}

		conffloat, _ := CheckStore(c, "CONFFLOAT")
		if conffloat.Float() != 3.33333 {
			t.Errorf(`Store item value was not "3.33333", but was %s`, conffloat)
		}

		confbool, _ := CheckStore(c, "CONFBOOL")
		if confbool.Bool() != true {
			t.Errorf(`Store item value was not "true", but was %t`, confbool)
		}

		confsect, _ := CheckStore(c, "SECTION_BLUE")
		if confsect.Value != "bondi" {
			t.Errorf(`Store item value was not "bondi", but was %s`, confsect)
		}

		conflist, _ := CheckStore(c, "CONFLIST")
		have := strings.Join(conflist.List("e"), ",")
		expected := strings.Join([]string{"a", "b", "c", "d", "e"}, ",")
		if bytes.Compare([]byte(have), []byte(expected)) != 0 {
			t.Errorf(`Store item value was not [a,b,c,d,e], but was [%s]`, have)
		}

		confasset, exists := CheckStore(c, "CONFASSET")
		if confasset.Value != "FROM_ASSET" {
			t.Errorf(`Store item value from assets configuration was not "FROM_ASSET", but was %s`, confasset)
		}
	}

	a := New("store", Mode("testing", true), WithAssets(TestAsset))

	a.Manage(NewRoute("GET", "/store", false, []Manage{storeroute}))

	a.LoadConfFile(testconffile())

	a.Configure()

	ac, _ := a.Env.Assets.GetByte("assets/bin.conf")
	a.LoadConfByte(ac, "bin.conf")

	p := NewPerformer(t, a, 200, "GET", "/store")

	performFor(p)
}
