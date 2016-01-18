package blueprint

import (
	"path/filepath"

	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/xrr"
)

// Routes returns a map of all routes attached to the app.
//func (a *App) Routes() Routes {
//	allroutes := make(Routes)
//	for _, blueprint := range a.Blueprints() {
//		for _, route := range blueprint.Routes {
//			allroutes[route.Name()] = route
//		}
//	}
//	return allroutes
//}

type Blueprints interface {
	Blueprint
	ListBlueprints() []Blueprint
	BlueprintExists(string) (Blueprint, bool)
	Attach(...Blueprint)
	Mount(string, ...Blueprint) error
}

type blueprints struct {
	Blueprint
}

func NewBlueprints(prefix string, fn HandleFn, mk state.Make) Blueprints {
	return &blueprints{
		Blueprint: newBlueprint(prefix, NewHandles(fn), NewMakes(mk)),
	}
}

func (b *blueprints) ListBlueprints() []Blueprint {
	type IterC func(bs []Blueprint, fn IterC)

	var bps []Blueprint

	bps = append(bps, b)

	iter := func(bs []Blueprint, fn IterC) {
		for _, x := range bs {
			bps = append(bps, x)
			fn(x.Descendents(), fn)
		}
	}

	iter(b.Descendents(), iter)

	return bps
}

func (b *blueprints) BlueprintExists(prefix string) (Blueprint, bool) {
	for _, bp := range b.ListBlueprints() {
		if bp.Prefix() == prefix {
			return bp, true
		}
	}
	return nil, false
}

// Given any number of Blueprints, RegisterBlueprints registers each with the App.
func (b *blueprints) Attach(blueprints ...Blueprint) {
	for _, blueprint := range blueprints {
		existing, exists := b.BlueprintExists(blueprint.Prefix())
		if !exists {
			blueprint.Register()
			b.Parent(blueprint)
		}
		if exists {
			for _, rt := range blueprint.Held() {
				existing.Manage(rt)
			}
		}
	}
}

var AlreadyRegistered = xrr.NewXrror("only unregistered blueprints may be mounted; %s is already registered").Out

// Mount attaches each provided Blueprint to the given string mount point.
func (b *blueprints) Mount(point string, blueprints ...Blueprint) error {
	var bs []Blueprint
	for _, blueprint := range blueprints {
		if blueprint.Registered() {
			return AlreadyRegistered(blueprint.Prefix)
		}

		newPrefix := filepath.ToSlash(filepath.Join(point, blueprint.Prefix()))

		nbp := newBlueprint(newPrefix, blueprint, blueprint)

		nbp.managers = combineManagers(b, blueprint.Managers())

		for _, rt := range blueprint.Held() {
			nbp.Manage(rt)
		}

		bs = append(bs, nbp)
	}
	b.Attach(bs...)
	return nil
}
