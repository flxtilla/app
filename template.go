package flotilla

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/thrisp/djinn"
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

func (env *Env) TemplatorInit() {
	if env.Templator == nil {
		env.Templator = NewTemplator(env)
	}
}

func (env *Env) TemplateDirs(dirs ...string) []string {
	storedirs := env.Store["TEMPLATE_DIRECTORIES"].List(dirs...)
	if env.Templator != nil {
		env.Templator.UpdateTemplateDirs(storedirs...)
		return env.Templator.ListTemplateDirs()
	}
	return storedirs
}

func NewTemplator(env *Env) *templator {
	j := &templator{Djinn: djinn.Empty()}
	j.UpdateTemplateDirs(env.Store["TEMPLATE_DIRECTORIES"].List()...)
	j.SetConf(djinn.Loaders(NewLoader(env)), djinn.TemplateFunctions(env.tplfunctions))
	return j
}

func (t *templator) ListTemplateDirs() []string {
	return t.TemplateDirs
}

func (t *templator) ListTemplates() []string {
	var ret []string
	for _, l := range t.Djinn.Loaders {
		ts := l.ListTemplates().([]string)
		ret = append(ret, ts...)
	}
	return ret
}

func (t *templator) UpdateTemplateDirs(dirs ...string) {
	for _, dir := range dirs {
		t.TemplateDirs = doAdd(dir, t.TemplateDirs)
	}
}

func NewLoader(env *Env) *Loader {
	fl := &Loader{env: env, FileExtensions: []string{".html", ".dji"}}
	return fl
}

func (fl *Loader) ValidFileExtension(ext string) bool {
	for _, extension := range fl.FileExtensions {
		if extension == ext {
			return true
		}
	}
	return false
}

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
	return "", newError("Template %s does not exist", name)
}
