package flotilla

import "testing"

var fauxconf bool

func FauxConf() Configuration {
	return func(a *App) error {
		fauxconf = true
		return nil
	}
}

func TestConfiguration(t *testing.T) {
	a := testApp(
		t,
		"configurationTest",
		Mode("prodnnuction", true),
		EnvItem("value:set", "section_value:set"),
		FauxConf(),
	)

	var testmodes *Modes

	var envval, envsectionval *StoreItem

	var non bool

	confTester := func(t *testing.T) Manage {
		return func(c Ctx) {
			tm, _ := c.Call("mode")
			testmodes = tm.(*Modes)
			if !testmodes.Development || !testmodes.Testing || testmodes.Production {
				t.Errorf("Modes not properly set: %+v", testmodes)
			}
			envval, _ = CheckStore(c, "VALUE")
			if envval.Value != "set" {
				t.Errorf(`EnvItem was not set properly; was EnvItem("value:set"), but retrieved Store value is %s`, envval)
			}
			envsectionval, _ = CheckStore(c, "SECTION_VALUE")
			if envsectionval.Value != "set" {
				t.Errorf(`EnvItem was not set properly; was EnvItem("section_value:set"), but retrieved Store value is %s`, envsectionval)
			}
			_, non = CheckStore(c, "NON")
			if non {
				t.Errorf(`A value was found in the store that should not exist.`)
			}

		}
	}

	exp, _ := NewExpectation(200, "GET", "/configuration", confTester)

	SimplePerformer(t, a, exp).Perform()

	if !fauxconf {
		t.Errorf(`Arbitrary Configuration function FauxConf was not properly set or used.`)
	}
}
