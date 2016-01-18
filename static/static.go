package static

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/thrisp/flotilla/assets"
	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/store"
)

type Static interface {
	Staticr
	SwapStaticr(Staticr)
}

type static struct {
	Staticr
}

func New(s store.Store, a assets.Assets) Static {
	return &static{
		Staticr: defaultStaticr(s, a),
	}
}

func (s *static) SwapStaticr(st Staticr) {
	s.Staticr = st
}

type Staticr interface {
	//assets.Assets
	//store.Store
	StaticDirs(...string) []string
	Exists(state.State, string) bool
	ManageFn(state.State)
}

type staticr struct {
	s store.Store
	a assets.Assets
}

func defaultStaticr(s store.Store, a assets.Assets) Staticr {
	return &staticr{
		s: s,
		a: a,
	}
}

func (st *staticr) StaticDirs(added ...string) []string {
	dirs := st.s.List("STATIC_DIRECTORIES")
	if added != nil {
		newDirs := dirs
		for _, _ = range dirs {
			//newDirs = doAdd(dir, newDirs)
		}
		st.s.Add("STATIC_DIRECTORIES", strings.Join(newDirs, ","))
	}
	return dirs
}

func (st *staticr) appStaticFile(requested string, s state.State) bool {
	exists := false
	for _, dir := range st.s.List("static_directories") {
		filepath.Walk(dir, func(path string, _ os.FileInfo, _ error) (err error) {
			if filepath.Base(path) == requested {
				f, _ := os.Open(path)
				serveStatic(s, f)
				exists = true
			}
			return err
		})
	}
	return exists
}

func (st *staticr) appAssetFile(requested string, s state.State) bool {
	exists := false
	f, err := st.a.GetAsset(requested)
	if err == nil {
		serveStatic(s, f)
		exists = true
	}
	return exists
}

func (st *staticr) Exists(s state.State, requested string) bool {
	exists := st.appStaticFile(requested, s)
	if !exists {
		exists = st.appAssetFile(requested, s)
	}
	return exists
}

func (st *staticr) ManageFn(s state.State) {
	if !st.Exists(s, requestedFile(s)) {
		abortStatic(s)
	} else {
		s.Call("header_now")
	}
}

func requestedFile(s state.State) string {
	rq := s.Request()
	return filepath.Base(rq.URL.Path)
}

func abortStatic(s state.State) {
	s.Call("abort", 404)
}

func serveStatic(s state.State, f http.File) {
	s.Call("serve_file", f)
}
