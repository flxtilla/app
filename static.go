package flotilla

import (
	"net/http"
	"os"
	"path/filepath"
)

type (
	Staticor interface {
		StaticDirs(...string) []string
		Exists(Ctx, string) bool
		Manage(Ctx)
	}

	staticor struct {
		app        *App
		staticDirs []string
	}
)

func StaticorInit(a *App) {
	if a.Env.Staticor == nil {
		a.Env.Staticor = NewStaticor(a)
	}
}

func NewStaticor(a *App) *staticor {
	s := &staticor{app: a}
	s.StaticDirs(s.app.Env.Store["STATIC_DIRECTORIES"].List()...)
	return s
}

func (env *Env) StaticDirs(dirs ...string) []string {
	storedirs := env.Store["STATIC_DIRECTORIES"].List(dirs...)
	if env.Staticor != nil {
		return env.Staticor.StaticDirs(storedirs...)
	}
	return storedirs
}

func (s *staticor) StaticDirs(dirs ...string) []string {
	for _, dir := range dirs {
		s.staticDirs = doAdd(dir, s.staticDirs)
	}
	return s.staticDirs
}

func (s *staticor) appStaticFile(requested string, c Ctx) bool {
	exists := false
	for _, dir := range s.app.StaticDirs() {
		filepath.Walk(dir, func(path string, _ os.FileInfo, _ error) (err error) {
			if filepath.Base(path) == requested {
				f, _ := os.Open(path)
				servestatic(c, f)
				exists = true
			}
			return err
		})
	}
	return exists
}

func (s *staticor) appAssetFile(requested string, c Ctx) bool {
	exists := false
	f, err := s.app.Assets.Get(requested)
	if err == nil {
		servestatic(c, f)
		exists = true
	}
	return exists
}

func (s *staticor) Exists(c Ctx, requested string) bool {
	exists := s.appStaticFile(requested, c)
	if !exists {
		exists = s.appAssetFile(requested, c)
	}
	return exists
}

func (s *staticor) Manage(c Ctx) {
	if !s.Exists(c, requestedfile(c)) {
		abortstatic(c)
	} else {
		c.Call("headernow")
	}
}

func requestedfile(c Ctx) string {
	rq, _ := c.Call("request")
	return filepath.Base(rq.(*http.Request).URL.Path)
}

func abortstatic(c Ctx) {
	c.Call("abort", 404)
}

func servestatic(c Ctx, f http.File) {
	c.Call("servefile", f)
}
