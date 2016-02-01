package template

import (
	"io"
	"strings"

	"github.com/thrisp/djinn"
	"github.com/thrisp/flotilla/asset"
	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/store"
	"github.com/thrisp/flotilla/xrr"
)

type Templatr interface {
	TemplateDirs(...string) []string
	ListTemplates() []string
	AddTemplateFunctions(fns map[string]interface{}) error
	SetTemplateFunctions()
	Render(io.Writer, string, interface{}) error
	RenderTemplate(state.State, string, interface{}) error
}

type templatr struct {
	*djinn.Djinn
	s store.Store
	a asset.Assets
	*functions
}

func DefaultTemplatr(s store.Store, a asset.Assets) Templatr {
	t := &templatr{
		Djinn: djinn.Empty(),
		s:     s,
		a:     a,
	}
	t.functions = &functions{t, false, nil}
	t.SetConf(djinn.Loaders(newLoader(t)))
	return t
}

func doAdd(s string, ss []string) []string {
	if isAppendable(s, ss) {
		ss = append(ss, s)
	}
	return ss
}

func isAppendable(s string, ss []string) bool {
	for _, x := range ss {
		if x == s {
			return false
		}
	}
	return true
}

func (t *templatr) TemplateDirs(added ...string) []string {
	dirs := t.s.List("TEMPLATE_DIRECTORIES")
	if added != nil {
		for _, add := range added {
			dirs = doAdd(add, dirs)
		}
		t.s.Add("TEMPLATE_DIRECTORIES", strings.Join(dirs, ","))
	}
	return dirs
}

func (t *templatr) ListTemplates() []string {
	var ret []string
	for _, l := range t.Djinn.Loaders {
		ts := l.ListTemplates()
		ret = append(ret, ts...)
	}
	return ret
}

type functions struct {
	t                 *templatr
	set               bool
	templateFunctions map[string]interface{}
}

func addTplFunc(f *functions, name string, fn interface{}) {
	if f.templateFunctions == nil {
		f.templateFunctions = make(map[string]interface{})
	}
	f.templateFunctions[name] = fn
}

func (f *functions) SetTemplateFunctions() {
	f.t.SetConf(djinn.TemplateFunctions(f.templateFunctions))
	f.set = true
}

var AlreadySet = xrr.NewXrror("cannot set template functions: already set")

func (f *functions) AddTemplateFunctions(fns map[string]interface{}) error {
	if !f.set {
		for k, v := range fns {
			addTplFunc(f, k, v)
		}
		return nil
	}
	return AlreadySet
}

func (t *templatr) RenderTemplate(s state.State, template string, data interface{}) error {
	s.Push(func(ps state.State) {
		td := NewTemplateData(ps, data)
		t.Render(ps.RWriter(), template, td)
	})
	return nil
}
