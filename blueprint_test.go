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

	f := New("flotilla_test_Blueprint")

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

	f.RegisterBlueprints(b)

	f.Configure(f.Configuration...)

	expected := "/blueprint/test_blueprint"

	p := NewPerformer(t, f, 200, method, expected)

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
	var passed1, passed2 bool
	f := New("flotilla_test_BlueprintMerge")
	m1 := func(c Ctx) { passed1 = true }
	m2 := func(c Ctx) { passed2 = true }
	b := NewBlueprint("/blueprint")
	one := NewRoute(method, "/one/:param", false, []Manage{m1})
	b.Manage(one)
	c := NewBlueprint("/blueprint")
	two := NewRoute(method, "/two/:param", false, []Manage{m2})
	c.Manage(two)
	f.RegisterBlueprints(b, c)
	f.Configure(f.Configuration...)
	p1 := NewPerformer(t, f, 200, method, "/blueprint/one/test")
	performFor(p1)
	p2 := NewPerformer(t, f, 200, method, "/blueprint/two/test")
	performFor(p2)
	if passed1 != true && passed2 != true {
		t.Errorf("Blueprint routes were not merged properly.")
	}
}

func TestBlueprintRegister(t *testing.T) {
	for _, m := range METHODS {
		registerBlueprints(m, t)
	}
}

func mountBlueprint(method string, t *testing.T) {
	var passed bool

	f := New("flotilla_test_BlueprintMount")

	b := NewBlueprint("/mount")

	m := func(c Ctx) { passed = true }

	one := NewRoute(method, "/mounted/1", false, []Manage{m})

	two := NewRoute(method, "/mounted/2", false, []Manage{m})

	b.Manage(one)

	b.Manage(two)

	f.Mount("/test/one", b)

	f.Mount("/test/two", b)

	f.RegisterBlueprints(b)

	f.Configure(f.Configuration...)

	err := f.Mount("/cannot", b)

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

	perform("/mount/mounted/1", method, f)
	perform("/mount/mounted/2", method, f)
	perform("/test/one/mount/mounted/1", method, f)
	perform("/test/two/mount/mounted/1", method, f)
	perform("/test/one/mount/mounted/2", method, f)
	perform("/test/two/mount/mounted/2", method, f)
}

func TestMountBlueprint(t *testing.T) {
	for _, m := range METHODS {
		mountBlueprint(m, t)
	}
}
