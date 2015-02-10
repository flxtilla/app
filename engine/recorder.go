package engine

import (
	"net/http"
	"time"
)

type (
	Recorder struct {
		start     time.Time
		stop      time.Time
		latency   time.Duration
		status    int
		method    string
		path      string
		requester string
	}
)

func newRecorder() *Recorder {
	return &Recorder{start: time.Now()}
}

func (r *Recorder) StopRecorder() {
	r.stop = time.Now()
}

func (r *Recorder) Requester(req *http.Request) {
	//rqstr := req.Header.Get("X-Real-IP")
	//if len(rqstr) == 0 {
	//	rqstr = req.Header.Get("X-Forwarded-For")
	//}
	//if len(rqstr) == 0 {
	//	rqstr = req.RemoteAddr
	//}
	r.requester = req.RemoteAddr //rqstr
}

func (r *Recorder) Latency() time.Duration {
	return r.stop.Sub(r.start)
}

func (r *Recorder) PostProcess(req *http.Request, withstatus int) {
	r.StopRecorder()
	r.stop = time.Now()
	r.latency = r.Latency()
	r.Requester(req)
	r.method = req.Method
	r.path = req.URL.Path
	r.status = withstatus
}

/*
func (r *Recorder) Fmt() string {
	return fmt.Sprintf("%s	%s	%s	%3d	%s	%s	%s", r.start, r.stop, r.latency, r.status, r.method, r.path, r.requester)
}

func (r *Recorder) LogFmt() string {
	return fmt.Sprintf("%v |%s %3d %s| %12v | %s |%s %s %-7s %s",
		r.stop.Format("2006/01/02 - 15:04:05"),
		StatusColor(r.status), r.status, reset,
		r.latency,
		r.requester,
		MethodColor(r.method), reset, r.method,
		r.path)
}
*/
