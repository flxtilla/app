package flotilla

import (
	"fmt"
	"log"
	"os"
)

type (
	Signal []byte

	Signals chan Signal

	queue func(string)

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

func (m *Messaging) Out(message string) {
	m.Logger.Printf(" %s", message)
}

func (m *Messaging) Panic(message string) {
	log.Println(fmt.Errorf("[Flotilla Panic] %s", message))
	m.Signals <- FlotillaPanic
	m.Signals <- []byte(message)
}

func (m *Messaging) Emit(message string) {
	m.Signals <- []byte(message)
}

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
