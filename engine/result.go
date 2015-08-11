package engine

import (
	"net/http"
	"time"

	"github.com/thrisp/flotilla/xrr"
)

type Result struct {
	*Recorder
	xrr.Xrroror
	Code   int
	Rule   Rule
	Params Params
	TSR    bool
}

func NewResult(code int, rule Rule, params Params, tsr bool) *Result {
	return &Result{
		Recorder: newRecorder(),
		Xrroror:  xrr.NewXrroror(),
		Code:     code,
		Rule:     rule,
		Params:   params,
		TSR:      tsr,
	}
}

type Recorder struct {
	RStart     time.Time
	RStop      time.Time
	RLatency   time.Duration
	RStatus    int
	RMethod    string
	RPath      string
	RRequester string
}

func newRecorder() *Recorder {
	return &Recorder{RStart: time.Now()}
}

func (r *Recorder) StopRecorder() {
	r.RStop = time.Now()
}

func (r *Recorder) Latency() time.Duration {
	return r.RStop.Sub(r.RStart)
}

func (r *Recorder) PostProcess(req *http.Request, withstatus int) {
	r.StopRecorder()
	r.RLatency = r.Latency()
	r.RRequester = req.RemoteAddr
	r.RMethod = req.Method
	r.RPath = req.URL.Path
	r.RStatus = withstatus
}
