package response

import (
	"net/http"

	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/xrr"
)

func abort(s state.State, code int) error {
	if code >= 0 {
		w := s.RWriter()
		w.WriteHeader(code)
		w.WriteHeaderNow()
	}
	return nil
}

func headerNow(s state.State) error {
	s.RWriter().WriteHeaderNow()
	return nil
}

func headerWrite(s state.State, code int, values ...[]string) error {
	if code >= 0 {
		s.RWriter().WriteHeader(code)
	}
	headerModify(s, "set", values...)
	return nil
}

func headerModify(s state.State, action string, values ...[]string) error {
	w := s.RWriter()
	switch action {
	case "set":
		for _, v := range values {
			w.Header().Set(v[0], v[1])
		}
	default:
		for _, v := range values {
			w.Header().Add(v[0], v[1])
		}
	}
	return nil
}

func isWritten(s state.State) bool {
	return s.RWriter().Written()
}

var InvalidStatusCode = xrr.NewXrror("Cannot send a redirect with status code %d").Out

func release(s state.State) {
	w := s.RWriter()
	if !w.Written() {
		s.Out(s)
		s.SessionRelease(w)
	}
}

func redirect(s state.State, code int, location string) error {
	if code >= 300 && code <= 308 {
		s.Bounce(func(ps state.State) {
			release(ps)
			w := ps.RWriter()
			http.Redirect(w, ps.Request(), location, code)
			w.WriteHeaderNow()
		})
		return nil
	} else {
		return InvalidStatusCode(code)
	}
}

func servePlain(s state.State, code int, data string) error {
	s.Push(func(ps state.State) {
		headerWrite(ps, code, []string{"Content-Type", "text/plain"})
		ps.RWriter().Write([]byte(data))
	})
	return nil
}

func serveFile(s state.State, f http.File) error {
	fi, err := f.Stat()
	if err == nil {
		http.ServeContent(s.RWriter(), s.Request(), fi.Name(), fi.ModTime(), f)
	}
	return err
}

func writeToResponse(s state.State, data string) error {
	s.RWriter().Write([]byte(data))
	return nil
}
