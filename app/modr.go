package app

import (
	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/xrr"
)

// Modr is an interface for managing any number of modes.
type Modr interface {
	GetMode(string) bool
	SetMode(string, bool) error
}

type modr struct {
	Development bool
	Production  bool
	Testing     bool
}

// DefaultModr returns a default Modr to manage Development, Production, and
// Testing modes.
func DefaultModr() Modr {
	return &modr{true, false, false}
}

// GetMode returns a boolean value for the provided string mode, false if not
// an existing mode.
func (m *modr) GetMode(mode string) bool {
	switch mode {
	case "dev", "development":
		return m.Development
	case "production":
		return m.Production
	case "test", "testing":
		return m.Testing
	}
	return false
}

var setModeError = xrr.NewXrror("mode could not be set to %s").Out

// SetMode sets the Mode indicated with a string with the provided boolean value.
// e.g. env.SetMode("Production", true)
func (m *modr) SetMode(mode string, value bool) error {
	switch mode {
	case "Development":
		m.Development = value
		return nil
	case "Production":
		m.Production = value
		return nil
	case "Testing":
		m.Testing = value
		return nil
	}
	return setModeError(mode)
}

// Given State and a string denoting a Mode, ModeIs returns a boolean value
// for that mode. If mode string is does not exist, returns false.
func ModeIs(s state.State, is string) bool {
	m, _ := s.Call("mode_is", is)
	return m.(bool)
}
