package flash

import (
	//"github.com/thrisp/flotilla/extension"
	"github.com/thrisp/flotilla/session"
)

//var flashFns = []extension.Function{
//	mkFunction("flasher", flshr),
//	mkFunction("flash", flash),
//}

//var FlashFxtension extension.Fxtension = extension.New("Flash_Fxtension", flashFns...)

type Flashes map[string][]string

type Flasher interface {
	Write(string) []string
	WriteAll() Flashes
	In(session.SessionStore) bool
	Out(session.SessionStore) bool
	Flash(string, string)
}

func New() Flasher {
	return &flasher{}
}

type flasher struct {
	readOnce bool
	flashes  Flashes
}

func (f *flasher) Write(key string) []string {
	if ret, ok := f.flashes[key]; ok {
		f.readOnce = true
		return ret
	}
	return nil
}

func (f *flasher) WriteAll() Flashes {
	ret := make(Flashes)
	for k, v := range f.flashes {
		ret[k] = v
	}
	f.readOnce = true
	f.flashes = nil
	return ret
}

func (f *flasher) In(s session.SessionStore) bool {
	if in := s.Get("_flashes"); in != nil {
		if inf, ok := in.(Flashes); ok {
			f.flashes = inf
			return true
		}
	}
	return false
}

func (f *flasher) Out(s session.SessionStore) bool {
	if err := s.Set("_flashes", f.flashes); err != nil {
		return true
	}
	return false
}

func (f *flasher) Flash(key, value string) {
	if f.flashes == nil {
		f.flashes = make(Flashes)
	}
	f.flashes[key] = append(f.flashes[key], value)
}

//func Flshr(c Ctx) Flasher {
//	fl, _ := c.Call("flasher")
//	return fl.(Flasher)
//}

//func flshr(c *ctx) (Flasher, error) {
//	return c.Flasher, nil
//}

//func flash(c *ctx, category, value string) error {
//	c.Flasher.Flash(category, value)
//	return nil
//}
