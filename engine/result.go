package engine

import (
	"net/http"
	"time"

	"github.com/thrisp/flotilla/xrr"
)

type Result struct {
	*Recorder
	xrr.Erroror
	Code   int
	Rule   Rule
	Params Params
	TSR    bool
}

func NewResult(code int, rule Rule, params Params, tsr bool) *Result {
	return &Result{
		Recorder: newRecorder(),
		Erroror:  xrr.DefaultErroror(),
		Code:     code,
		Rule:     rule,
		Params:   params,
		TSR:      tsr,
	}
}

type Recorder struct {
	start     time.Time
	stop      time.Time
	latency   time.Duration
	status    int
	method    string
	path      string
	requester string
}

func newRecorder() *Recorder {
	return &Recorder{start: time.Now()}
}

func (r *Recorder) StopRecorder() {
	r.stop = time.Now()
}

func (r *Recorder) Latency() time.Duration {
	return r.stop.Sub(r.start)
}

func (r *Recorder) PostProcess(req *http.Request, withstatus int) {
	r.StopRecorder()
	r.stop = time.Now()
	r.latency = r.Latency()
	r.requester = req.RemoteAddr
	r.method = req.Method
	r.path = req.URL.Path
	r.status = withstatus
}
