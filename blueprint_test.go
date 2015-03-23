package flotilla

import (
	"fmt"
	"reflect"
	"testing"
)

func testBlueprint(method string, t *testing.T) {
	var passed bool
	var passone bool
	var passmultiple []bool
	var inc int
	var incis bool

	a := testApp(t, "testBlueprint", nil, nil)

	b := NewBlueprint("/blueprint")

	bm := func(c Ctx) {
		if inc == 1 {
			incis = true
			inc++
		} else {
			incis = false
		}
		passone = true
		passmultiple = append(passmultiple, true)
	}

	bm0 := func(c Ctx) {
		if inc == 0 {
			incis = true
			inc++
		} else {
			incis = false
		}
	}

	bm1 := func(c Ctx) {
		if inc == 2 {
			incis = true
			inc++
		} else {
			incis = false
		}
	}

	b.Use(bm, bm, bm)

	b.UseAt(0, bm0)

	b.UseAt(5, bm1)

	m := func(c Ctx) { passed = true }

	reflect.ValueOf(b).MethodByName(method).Call([]reflect.Value{reflect.ValueOf("/test_blueprint"), reflect.ValueOf(m)})

	a.RegisterBlueprints(b)

	a.Configure()

	expected := "/blueprint/test_blueprint"

	p := NewPerformer(t, a, 200, method, expected)

	performFor(p)

	if passed != true {
		t.Errorf("%s blueprint route: %s was not invoked.", method, expected)
	}

	if inc != 3 || !incis {
		t.Errorf("Error setting and cycling through blueprint level Manage functions: %+v", b)
	}

	if passone != true {
		t.Errorf("Blueprint level Manage %#v was not invoked: %t.", bm, passone)
	}

	if len(passmultiple) > 1 {
		t.Errorf("Blueprint level Manage %#v was duplicated.", bm)
	}

	if passmultiple[0] != true {
		t.Errorf("Blueprint level Manage %#v was used in error.", bm)
	}

}

func TestBlueprint(t *testing.T) {
	for _, m := range METHODS {
		testBlueprint(m, t)
	}
}

func registerBlueprints(method string, t *testing.T) {
	var passed0, passed1, passed2 bool
	a := testApp(t, "testRegisterBlueprints", nil, nil)
	m0 := func(c Ctx) { passed0 = true }
	m1 := func(c Ctx) { passed1 = true }
	m2 := func(c Ctx) { passed2 = true }
	b0 := NewBlueprint("/")
	zero := NewRoute(method, "/zero/:param", false, []Manage{m0})
	b0.Manage(zero)
	b1 := NewBlueprint("/blueprint")
	one := NewRoute(method, "/route/one", false, []Manage{m1})
	b1.Manage(one)
	b2 := NewBlueprint("/blueprint")
	two := NewRoute(method, "/route/two", false, []Manage{m2})
	b2.Manage(two)
	a.RegisterBlueprints(b0, b1, b2)
	a.Configure()
	p0 := NewPerformer(t, a, 200, method, "/zero/test")
	performFor(p0)
	p1 := NewPerformer(t, a, 200, method, "/blueprint/route/one")
	performFor(p1)
	p2 := NewPerformer(t, a, 200, method, "/blueprint/route/two")
	performFor(p2)
	if passed0 != true && passed1 != true && passed2 != true {
		t.Errorf("Blueprint routes were not merged properly.")
	}
	var paths []string
	for _, rt := range a.Routes() {
		paths = append(paths, rt.path)
	}
	for _, expected := range []string{"/zero/:param", "/blueprint/route/one", "/blueprint/route/two"} {
		if !existsIn(expected, paths) {
			t.Errorf("Expected route with path %s was not found in added routes.", expected)
		}
	}
}

func TestBlueprintRegister(t *testing.T) {
	for _, m := range METHODS {
		registerBlueprints(m, t)
	}
}

func chainBlueprints(method string, t *testing.T) {
	var x1, x2, x3 bool
	var y int
	a := testApp(t, "testChainedBlueprints", nil, nil)
	a.Use(func(c Ctx) {
		x1 = true
		y = 1
	})
	b := a.NewBlueprint("/blueprintone")
	b.Use(func(c Ctx) {
		if x1 {
			x2 = true
			y = 2
		}
	})
	c := b.NewBlueprint("/blueprinttwo")
	third := NewRoute(method, "/third", false, []Manage{func(c Ctx) {}})
	c.Manage(third)
	c.Use(func(c Ctx) {
		if x1 && x2 {
			x3 = true
			y = 3
		}
	})
	a.Configure()
	p := NewPerformer(t, a, 200, method, "/blueprintone/blueprinttwo/third")
	performFor(p)
	if !x1 && !x2 && !x3 && !(y == 3) {
		t.Errorf("Blueprint Manage chain error, chained test blueprint did not execute expected Manage.")
	}
}

func TestChainBlueprints(t *testing.T) {
	for _, m := range METHODS {
		chainBlueprints(m, t)
	}
}

func mountBlueprint(method string, t *testing.T) {
	var passed bool

	a := testApp(t, "testMountBlueprint", nil, nil)

	b := NewBlueprint("/mount")

	m := func(c Ctx) { passed = true }

	one := NewRoute(method, "/mounted/1", false, []Manage{m})

	two := NewRoute(method, "/mounted/2", false, []Manage{m})

	b.Manage(one)

	b.Manage(two)

	a.Mount("/test/one", b)

	a.Mount("/test/two", b)

	a.RegisterBlueprints(b)

	a.Configure()

	err := a.Mount("/cannot", b)

	if err == nil {
		t.Errorf("mounting a registered blueprint return no error")
	}

	perform := func(expected string, method string, app *App) {
		p := NewPerformer(t, app, 200, method, expected)

		performFor(p)

		if passed == false {
			t.Errorf(fmt.Sprintf("%s blueprint route: %s was not invoked.", method, expected))
		}

		passed = false
	}

	perform("/mount/mounted/1", method, a)
	perform("/mount/mounted/2", method, a)
	perform("/test/one/mount/mounted/1", method, a)
	perform("/test/two/mount/mounted/1", method, a)
	perform("/test/one/mount/mounted/2", method, a)
	perform("/test/two/mount/mounted/2", method, a)
}

func TestMountBlueprint(t *testing.T) {
	for _, m := range METHODS {
		mountBlueprint(m, t)
	}
}
