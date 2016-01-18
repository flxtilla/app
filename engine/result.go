package engine

import (
	"net/http"
	"time"

	"github.com/thrisp/flotilla/xrr"
)

type Result struct {
	Code   int
	Rule   Rule
	Params Params
	TSR    bool
	xrr.Xrroror
	Recorder
}

func NewResult(code int, rule Rule, params Params, tsr bool) *Result {
	return &Result{
		Code:     code,
		Rule:     rule,
		Params:   params,
		TSR:      tsr,
		Xrroror:  xrr.NewXrroror(),
		Recorder: newRecorder(),
	}
}

type Recorder interface {
	PostProcess(*http.Request, int)
	Record() *Recorded
}

type Recorded struct {
	Start, Stop             time.Time
	Latency                 time.Duration
	Status                  int
	Method, Path, Requester string
}

type recorder struct {
	*Recorded
}

func newRecorder() Recorder {
	return &recorder{&Recorded{Start: time.Now()}}
}

func (r *recorder) stopRecorder() {
	r.Stop = time.Now()
}

func (r *recorder) latency() time.Duration {
	return r.Stop.Sub(r.Start)
}

func (r *recorder) Record() *Recorded {
	return r.Recorded
}

func (r *recorder) PostProcess(req *http.Request, withstatus int) {
	r.stopRecorder()
	r.Latency = r.latency()
	r.Requester = req.RemoteAddr
	r.Method = req.Method
	r.Path = req.URL.Path
	r.Status = withstatus
}
