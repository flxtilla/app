package flotilla

import (
	"net/http"
	"os"
	"path/filepath"
)

type (
	// Staticor provides an interface to static files for an App.
	Staticor interface {
		// StaticDirs any number of strings and returns a string list for
		// adding to and listing the Staticor directories.
		StaticDirs(...string) []string

		// Exists takes a Ctx & string to determine if the Staticor can handle the
		// the file designated by the string.
		Exists(Ctx, string) bool

		// Manage is a flotilla.Manage function used by the Staticor.
		Manage(Ctx)
	}

	staticor struct {
		app        *App
		staticDirs []string
	}
)

// StaticorInit intializes a staticor from the Staticor provided to App.Env.
func StaticorInit(a *App) {
	if a.Env.Staticor == nil {
		a.Env.Staticor = NewStaticor(a)
	}
}

// NewStaticor returns a new default flotilla Staticor.
func NewStaticor(a *App) *staticor {
	s := &staticor{app: a}
	s.StaticDirs(s.app.Env.Store.List("STATIC_DIRECTORIES")...)
	return s
}

// StaticDirs takes any number of directories as string, adding them, and returning
// a string array of static directories in Env.Store
func (env *Env) StaticDirs(dirs ...string) []string {
	i, _ := env.Store.query("STATIC_DIRECTORIES")
	if i != nil {
		storedirs := i.List(dirs...)
		if env.Staticor != nil {
			return env.Staticor.StaticDirs(storedirs...)
		}
		return storedirs
	}
	return nil
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
