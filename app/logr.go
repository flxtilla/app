package app

import (
	"os"

	"github.com/thrisp/flotilla/log"
)

// Logr is an interface to logging that implements flotilla/log.Logger as well
// as swap method for changing the Logger when needed.
type Logr interface {
	log.Logger
	SwapLogger(log.Logger)
}

type defaultLogr struct {
	log.Logger
}

// SwapLogger changes the existing log.Logger to the provided log.Logger.
func (d *defaultLogr) SwapLogger(l log.Logger) {
	d.Logger = l
}

// DefaultLogr returns the default flotilla Logger.
func DefaultLogr() Logr {
	return &defaultLogr{
		Logger: log.New(os.Stdout, log.LInfo, log.DefaultTextFormatter()),
	}
}
