package flotilla

import (
	"fmt"
	"log"
	"os"
)

type (
	// Signal denotes a byte signal
	Signal []byte

	queue func(string)

	Signals struct {
		SGNL   chan Signal
		Queues map[string]queue
		Logger *log.Logger
	}
)

var (
	FlotillaPanic = []byte("flotilla-panic")
)

func newsignals() *Signals {
	ss := &Signals{}
	ss.Logger = log.New(os.Stdout, "[Flotilla]", 0)
	ss.Queues = ss.defaultqueues()
	ss.SGNL = make(chan Signal, 10000)
	return ss
}

func SignalQueue(a *App) {
	go func() {
		for sig := range a.SGNL {
			a.Out(fmt.Sprintf("%s", sig))
		}
	}()
}

// Message goes directly to a logger, if enabled.
func (s *Signals) Out(message string) {
	s.Logger.Printf(" %s", message)
}

// Panic emits a defined flotilla panic signal, send the message stdout, then
// emits the message as a byte to Signals.signals.
func (s *Signals) Panic(message string) {
	s.SGNL <- FlotillaPanic
	log.Println(fmt.Errorf("[FLOTILLA] %s", message))
	s.SGNL <- []byte(message)
}

// Emit takes a string that goes as []byte directly to engine.Signals
func (s *Signals) Emit(message string) {
	s.SGNL <- []byte(message)
}

// Send sends a message to the specified queue.
func (s *Signals) Send(queue string, message string) {
	go func() {
		if q, ok := s.Queues[queue]; ok {
			q(message)
		} else {
			s.Emit(message)
		}
	}()
}

func (s Signals) defaultqueues() map[string]queue {
	q := make(map[string]queue)
	q["out"] = s.Out
	q["panic"] = s.Panic
	q["emit"] = s.Emit
	return q
}
