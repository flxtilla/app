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
		testConf(
			Mode("prodnnuction", true),
			EnvItem("value:set", "section_value:set"),
			FauxConf(),
		),
		nil,
	)

	var testmodes *Modes

	var envval, envsectionval *StoreItem

	var non bool

	conffunc := func(c Ctx) {
		tm, _ := c.Call("mode")
		testmodes = tm.(*Modes)
		envval, _ = CheckStore(c, "VALUE")
		envsectionval, _ = CheckStore(c, "SECTION_VALUE")
		_, non = CheckStore(c, "NON")
	}

	a.GET("/configuration", conffunc)

	a.Configure(a.Configuration...)

	p := NewPerformer(t, a, 200, "GET", "/configuration")

	performFor(p)

	if !testmodes.Development || !testmodes.Testing || testmodes.Production {
		t.Errorf("Modes not properly set: %+v", testmodes)
	}

	if envval.Value != "set" {
		t.Errorf(`EnvItem was not set properly; was EnvItem("value:set"), but retrieved Store value is %s`, envval)
	}

	if envsectionval.Value != "set" {
		t.Errorf(`EnvItem was not set properly; was EnvItem("section_value:set"), but retrieved Store value is %s`, envsectionval)
	}

	if non {
		t.Errorf(`A value was found in the store that should not exist.`)
	}

	if !fauxconf {
		t.Errorf(`Arbitrary Configuration function FauxConf was not properly set or used.`)
	}
}
