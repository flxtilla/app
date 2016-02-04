package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thrisp/flotilla/app"
	"github.com/thrisp/flotilla/asset"
	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/store"
	"github.com/thrisp/flotilla/store/resources"
	"github.com/thrisp/flotilla/txst"
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

var TestAsset asset.AssetFS = asset.NewAssetFS(
	resources.Asset,
	resources.AssetDir,
	resources.AssetNames,
	"",
)

func testLocation() string {
	wd, _ := os.Getwd()
	ld, _ := filepath.Abs(wd)
	return ld
}

func testConfFile() string {
	return filepath.Join(testLocation(), "resources", "flotilla.conf")
}

type defaultreturns struct {
	i   int
	i64 int64
	f   float64
	b   bool
}

func CheckStore(s state.State, key string) (store.StoreItem, bool) {
	return nil, false
}

func TestStore(t *testing.T) {
	exp, _ := txst.NewExpectation(
		200,
		"GET",
		"/store",
		func(t *testing.T) state.Manage {
			return func(s state.State) {
				/*
					_, exists := CheckStore(s, "UNREAD_VALUE")
					if exists {
						t.Errorf(`Store item value exists, but should not.`)
					}

					confstr, _ := CheckStore(s, "CONFSTRING")
					if confstr.Value != "ONE" {
						t.Errorf(`Store item value was not "ONE", but was %s`, confstr.Value)
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

					confint, _ := CheckStore(s, "CONFINT")
					if confint.Int() != 2 {
						t.Errorf(`Store item value was not "ONE", but was %s`, confint)
					}

					confint64, _ := CheckStore(s, "CONFINT64")
					if confint64.Int() != 99999 {
						t.Errorf(`Store item value was not "99999", but was %s`, confint64)
					}

					conffloat, _ := CheckStore(s, "CONFFLOAT")
					if conffloat.Float() != 3.33333 {
						t.Errorf(`Store item value was not "3.33333", but was %s`, conffloat)
					}

					confbool, _ := CheckStore(s, "CONFBOOL")
					if confbool.Bool() != true {
						t.Errorf(`Store item value was not "true", but was %t`, confbool)
					}

					confsect, _ := CheckStore(s, "SECTION_BLUE")
					if confsect.Value != "bondi" {
						t.Errorf(`Store item value was not "bondi", but was %s`, confsect)
					}

					conflist, _ := CheckStore(s, "CONFLIST")
					have := strings.Join(conflist.List("e"), ",")
					expected := strings.Join([]string{"a", "b", "c", "d", "e"}, ",")
					if bytes.Compare([]byte(have), []byte(expected)) != 0 {
						t.Errorf(`Store item value was not [a,b,c,d,e], but was [%s]`, have)
					}

					confasset, exists := CheckStore(s, "CONFASSET")
					if confasset.Value != "FROM_ASSET" {
						t.Errorf(`Store item value from assets configuration was not "FROM_ASSET", but was %s`, confasset)
					}
				*/
			}
		})

	a := AppForTest(
		t,
		"testStore",
		app.Assets(TestAsset),
	)
	a.Load(testConfFile())
	ac, _ := a.GetAssetByte("assets/bin.conf")
	a.LoadByte(ac, "bin.conf")
	a.Configure()

	txst.SimplePerformer(t, a, exp).Perform()
}
