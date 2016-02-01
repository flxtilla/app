package session

import (
	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/state"
)

func deleteSession(s state.State, key string) error {
	return s.Delete(key)
}

func getSession(s state.State, key string) interface{} {
	return s.Get(key)
}

func SessionStore(s state.State) session.SessionStore {
	ss, _ := s.Call("session")
	return ss.(session.SessionStore)
}

func returnSession(s state.State) session.SessionStore {
	return s
}

func setSession(s state.State, key string, value interface{}) error {
	return s.Set(key, value)
}
