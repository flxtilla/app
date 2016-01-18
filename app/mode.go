package app

import (
	"github.com/thrisp/flotilla/state"
	"github.com/thrisp/flotilla/xrr"
)

// Modes configure specific modes to reference or differentate fuctionality.
// Unless set, an App defaults to Development true, Production false, and
// Testing false.
type Modes interface {
	GetMode(string) bool
	SetMode(string, bool) error
}

type modes struct {
	Development bool
	Production  bool
	Testing     bool
}

func defaultModes() Modes {
	return &modes{true, false, false}
}

func (m *modes) GetMode(mode string) bool {
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

var SetModeError = xrr.NewXrror("mode could not be set to %s").Out

// SetMode sets the Mode indicated with a string with the provided boolean value.
// e.g. env.SetMode("Production", true)
func (m *modes) SetMode(mode string, value bool) error {
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
	return SetModeError(mode)
}

func ModeIs(s state.State, is string) bool {
	m, _ := s.Call("mode_is", is)
	return m.(bool)
}
