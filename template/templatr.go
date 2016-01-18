package template

import (
	"io"
	"reflect"

	"github.com/thrisp/djinn"
	"github.com/thrisp/flotilla/assets"
	"github.com/thrisp/flotilla/store"
	"github.com/thrisp/flotilla/xrr"
)

type Templatr interface {
	Render(io.Writer, string, interface{}) error
	ListTemplates() []string
	AddTemplateFunctions(fns map[string]interface{}) error
	SetTemplateFunctions()
	AddContextProcessors(fns map[string]interface{})
}

type templatr struct {
	*djinn.Djinn
	s store.Store
	a assets.Assets
	*functions
	*processors
}

func DefaultTemplatr(s store.Store, a assets.Assets) Templatr {
	t := &templatr{
		Djinn:      djinn.Empty(),
		s:          s,
		a:          a,
		processors: &processors{},
	}
	t.functions = &functions{t, false, nil}
	t.SetConf(djinn.Loaders(newLoader(t)))
	return t
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

type processors struct {
	contextProcessors map[string]reflect.Value
}

func addContextProcessor(p *processors, name string, fn interface{}) {
	if p.contextProcessors == nil {
		p.contextProcessors = make(map[string]reflect.Value)
	}
	p.contextProcessors[name] = valueFunc(fn)
}

func (p *processors) AddContextProcessors(fns map[string]interface{}) {
	for k, v := range fns {
		addContextProcessor(p, k, v)
	}
}

func (t *templatr) ListTemplates() []string {
	var ret []string
	for _, l := range t.Djinn.Loaders {
		ts := l.ListTemplates()
		ret = append(ret, ts...)
	}
	return ret
}
