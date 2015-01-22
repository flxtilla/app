package flotilla

import (
	"fmt"
	"testing"
)

func testSignal(method string, t *testing.T) {
	var sent bool = false
	f := New("signals_test")
	testsignalq := func() {
		go func() {
			for msg := range f.Signals {
				m := fmt.Sprintf("%s", msg)
				if m != "TEST" {
					t.Errorf(fmt.Sprintf("Read signal is  %s not `TEST`", msg))
				}
				//fmt.Printf("test: %s\n", m)
			}
		}()
	}
	testsignalq()
	testqueue := func(s string) {
		if s != "SENT" {
			t.Errorf("Read signal is not `SENT`")
		}
	}
	f.Queues["testqueue"] = testqueue
	rt := NewRoute(method, "/test_signal_sent", false, []Manage{func(c *Ctx) {
		sent = true
		f.Emit("TEST")
		for i := 0; i < 10; i++ {
			f.Send("testqueue", "SENT")
		}
	}})
	f.Handle(rt)
	f.Configure(f.Configuration...)
	PerformRequest(f, method, "/test_signal_sent")
	if sent == false {
		t.Errorf("Signal handler was not invoked.")
	}
}

func TestSignal(t *testing.T) {
	for _, m := range METHODS {
		testSignal(m, t)
	}
}
