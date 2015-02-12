package flotilla

import (
	"fmt"
	"log"
	"os"
)

type (
	// A byte signal for App messaging.
	Signal []byte

	// A Signal channel for App messaging.
	Signals chan Signal

	queue func(string)

	// Messaging encapsulates signalling & logging in an App.
	Messaging struct {
		Signals chan Signal
		Queues  map[string]queue
		Logger  *log.Logger
	}
)

var (
	FlotillaPanic = []byte("flotilla-panic")
)

func newMessaging() *Messaging {
	m := &Messaging{}
	m.Logger = log.New(os.Stdout, "[Flotilla]", 0)
	m.Queues = m.defaultqueues()
	m.Signals = make(Signals, 100)
	return m
}

// Out sends the provided string to messaging logger.
func (m *Messaging) Out(message string) {
	m.Logger.Printf(" %s", message)
}

// Panic immediately logs the provided string, ans sends a FlotillaPanic signal
// and the message to messging Signals.
func (m *Messaging) Panic(message string) {
	log.Println(fmt.Errorf("[Flotilla Panic] %s", message))
	m.Signals <- FlotillaPanic
	m.Signals <- []byte(message)
}

// Emit send thes the provided message as a Signal to messaging Signals channel.
func (m *Messaging) Emit(message string) {
	m.Signals <- []byte(message)
}

// Send sends the message to the provided queue, with a fall through to Emit if
// the queue does not exist.
func (m *Messaging) Send(queue string, message string) {
	if q, ok := m.Queues[queue]; ok {
		q(message)
		return
	}
	m.Emit(message)
}

func (m Messaging) defaultqueues() map[string]queue {
	q := make(map[string]queue)
	q["out"] = m.Out
	q["panic"] = m.Panic
	q["emit"] = m.Emit
	return q
}
