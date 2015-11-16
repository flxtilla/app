package flotilla

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/thrisp/djinn"
	"github.com/thrisp/flotilla/xrr"
)

type (
	Templator interface {
		Render(io.Writer, string, interface{}) error
		ListTemplateDirs() []string
		ListTemplates() []string
		UpdateTemplateDirs(...string)
	}

	templator struct {
		*djinn.Djinn
		TemplateDirs []string
	}

	Loader struct {
		env            *Env
		FileExtensions []string
	}
)

// TemplatorInit intializes the default Templator if none is listed with the Env.
func TemplatorInit(a *App) {
	if a.Env.Templator == nil {
		a.Env.Templator = NewTemplator(a.Env)
	}
}

// TemplateDirs updates with the provided and returns existing template
// directories listed in the Env.
func (env *Env) TemplateDirs(dirs ...string) []string {
	i, _ := env.Store.query("TEMPLATE_DIRECTORIES")
	if i != nil {
		storedirs := i.List(dirs...)
		if env.Templator != nil {
			env.Templator.UpdateTemplateDirs(storedirs...)
			return env.Templator.ListTemplateDirs()
		}
		return storedirs
	}
	return nil
}

// NewTemplator returns a new default templator.
func NewTemplator(env *Env) *templator {
	j := &templator{Djinn: djinn.Empty()}
	j.UpdateTemplateDirs(env.Store.List("TEMPLATE_DIRECTORIES")...)
	j.SetConf(djinn.Loaders(NewLoader(env)), djinn.TemplateFunctions(env.tplfunctions))
	return j
}

// ListTemplateDirs lists template directories attached to the templator.
func (t *templator) ListTemplateDirs() []string {
	return t.TemplateDirs
}

// ListTemplates returns list of templates in hte Templator loaders.
func (t *templator) ListTemplates() []string {
	var ret []string
	for _, l := range t.Djinn.Loaders {
		ts := l.ListTemplates().([]string)
		ret = append(ret, ts...)
	}
	return ret
}

// UpdateTemplateDirs updatese the templator with the provided string directories.
func (t *templator) UpdateTemplateDirs(dirs ...string) {
	for _, dir := range dirs {
		t.TemplateDirs = doAdd(dir, t.TemplateDirs)
	}
}

// newLoader returns a new flotilla Loader.
func NewLoader(env *Env) *Loader {
	fl := &Loader{env: env, FileExtensions: []string{".html", ".dji"}}
	return fl
}

// ValidFileExtension returns a boolean for extension provided if the flotilla
// Loader allows the extension type.
func (fl *Loader) ValidFileExtension(ext string) bool {
	for _, extension := range fl.FileExtensions {
		if extension == ext {
			return true
		}
	}
	return false
}

// AssetTemplates returns templates linked to Assets in the flotilla Loader.
func (fl *Loader) AssetTemplates() []string {
	var ret []string
	for _, assetfs := range fl.env.Assets {
		for _, f := range assetfs.AssetNames() {
			if fl.ValidFileExtension(filepath.Ext(f)) {
				ret = append(ret, f)
			}
		}
	}
	return ret
}

// ListTemplates  lists templates in the flotilla Loader.
func (fl *Loader) ListTemplates() interface{} {
	var ret []string
	for _, p := range fl.env.TemplateDirs() {
		files, _ := ioutil.ReadDir(p)
		for _, f := range files {
			if fl.ValidFileExtension(filepath.Ext(f.Name())) {
				ret = append(ret, fmt.Sprintf("%s/%s", p, f.Name()))
			}
		}
	}
	ret = append(ret, fl.AssetTemplates()...)
	return ret
}

var TemplateDoesNotExist = xrr.NewXrror("Template %s does not exist.").Out

// Load a template by string name from the flotilla Loader.
func (fl *Loader) Load(name string) (string, error) {
	for _, p := range fl.env.TemplateDirs() {
		f := filepath.Join(p, name)
		if fl.ValidFileExtension(filepath.Ext(f)) {
			// existing template dirs
			if _, err := os.Stat(f); err == nil {
				file, err := os.Open(f)
				r, err := ioutil.ReadAll(file)
				return string(r), err
			}
			// binary assets
			if r, err := fl.env.Assets.Get(name); err == nil {
				r, err := ioutil.ReadAll(r)
				return string(r), err
			}
		}
	}
	return "", TemplateDoesNotExist(name)
}
