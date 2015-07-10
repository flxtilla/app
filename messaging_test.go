package flotilla

import (
	"fmt"
	"testing"
)

func testout(t *testing.T, a *App) Queue {
	return func(message string) {}
}

func testpanicq(t *testing.T, a *App) Queue {
	return func(message string) {
		a.Signals <- FlotillaPanic
		a.Signals <- []byte(message)
	}
}

func mkTestQueues(t *testing.T, a *App) {
	a.Messaging.Queues["out"] = testout(t, a)
	a.Messaging.Queues["panic"] = testpanicq(t, a)
}

func testsignalq(t *testing.T, a *App, against func(*testing.T, Signal)) {
	go func() {
		for msg := range a.Signals {
			against(t, msg)
		}
	}()
}

func testSignal(method string, t *testing.T) {
	a := testApp(t, "testSignal", WithQueue("none", func(string) {}))
	testsignalq(t, a, func(t *testing.T, msg Signal) {
		m := fmt.Sprintf("%s", msg)
		if m != "TEST" {
			t.Errorf(fmt.Sprintf("Read signal is  %s not `TEST`", msg))
		}
	})
	testqueue := func(s string) {
		if s != "SENT" {
			t.Errorf("Read signal is not `SENT`")
		}
	}
	a.Queues["testqueue"] = testqueue
	a.Manage(NewRoute(defaultRouteConf(method, "/test_signal_sent", []Manage{func(c Ctx) {
		a.Emit("TEST")
		for i := 0; i < 10; i++ {
			a.Send("testqueue", "SENT")
			a.Send("notaqueue", "TEST")
		}
	}})))
	ZeroExpectationPerformer(t, a, 200, method, "/test_signal_sent").Perform()
}

func TestSignal(t *testing.T) {
	for _, m := range METHODS {
		testSignal(m, t)
	}
}
