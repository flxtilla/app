package app

import (
	"testing"

	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/txst"
)

var fauxconf bool

func FauxConf() ConfigurationFn {
	return func(a *App) error {
		fauxconf = true
		return nil
	}
}

func TestConfiguration(t *testing.T) {
	a := AppForTest(
		t,
		"configurationTest",
		Mode("prodnnuction", true),
		Store("value:set", "section_value:set"),
		FauxConf(),
	)

	//var testmodes *Modes

	//var envval, envsectionval store.StoreItem

	//var non bool

	confTester := func(t *testing.T) state.Manage {
		return func(s state.State) {
			//tm, _ := c.Call("mode")
			//testmodes = tm.(*Modes)
			//if !testmodes.Development || !testmodes.Testing || testmodes.Production {
			//	t.Errorf("Modes not properly set: %+v", testmodes)
			//}
			//envval, _ = CheckStore(c, "VALUE")
			//if envval.Value != "set" {
			//	t.Errorf(`EnvItem was not set properly; was EnvItem("value:set"), but retrieved Store value is %s`, envval)
			//}
			//envsectionval, _ = CheckStore(c, "SECTION_VALUE")
			//if envsectionval.Value != "set" {
			//	t.Errorf(`EnvItem was not set properly; was EnvItem("section_value:set"), but retrieved Store value is %s`, envsectionval)
			//}
			//_, non = CheckStore(c, "NON")
			//if non {
			//	t.Errorf(`A value was found in the store that should not exist.`)
			//}
		}
	}

	exp, _ := txst.NewExpectation(200, "GET", "/configuration", confTester)

	txst.SimplePerformer(t, a, exp).Perform()

	//if !fauxconf {
	//	t.Errorf(`Arbitrary Configuration function FauxConf was not properly set or used.`)
	//}
}
