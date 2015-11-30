package flotilla

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// A byte signal for App messaging.
type Signal []byte

// A Signal channel for App messaging.
type Signals chan Signal

// A function taking any string.
type Queue func(string)

// Messaging encapsulates signalling & logging in an App.
type Messaging struct {
	Signals chan Signal
	Queues  map[string]Queue
	Logger  *log.Logger
}

var FlotillaPanic = []byte("flotilla-panic")

func newMessaging() *Messaging {
	m := &Messaging{}
	m.Logger = log.New(os.Stdout, "[Flotilla]", 0)
	m.Queues = m.defaultQueues()
	m.Signals = make(Signals, 100)
	return m
}

func flush(c Signals, m Signal) {
	c <- m
}

func (m *Messaging) Out(message string) {
	m.Send("out", message)
}

// Out sends the provided string to messaging logger.
func (m *Messaging) DefaultOut(message string) {
	m.Logger.Printf(" %s", message)
}

// Panic immediately logs the provided string, and sends a FlotillaPanic signal
// and the message to messaging Signals.
func (m *Messaging) Panic(message string) {
	m.Send("panic", message)
}

func (m *Messaging) DefaultPanic(message string) {
	log.Println(fmt.Errorf("[Flotilla Panic] %s", message))
}

// Emit send the provided message as a Signal to messaging Signals channel.
func (m *Messaging) Emit(message string) {
	m.Signals <- []byte(message)
}

var wg sync.WaitGroup

func exq(q Queue, m string) {
	q(m)
	wg.Done()
}

// Send sends the message to the provided queue, with a fall through to Emit if
// the queue does not exist.
func (m *Messaging) Send(queue string, message string) {
	q, ok := m.Queues[queue]
	wg.Add(1)
	if ok {
		go exq(q, message) //q(message)
	}
	if !ok {
		go exq(m.Emit, message) //m.Emit(message)
	}
	wg.Wait()
}

func (m Messaging) defaultQueues() map[string]Queue {
	return map[string]Queue{
		"out":   m.DefaultOut,
		"panic": m.DefaultPanic,
		"emit":  m.Emit,
	}
}
