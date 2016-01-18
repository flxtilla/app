package status

import "github.com/thrisp/flotilla/state"

type Statuses interface {
	GetStatus(int) Status
	SetRawStatus(int, ...state.Manage)
	SetStatus(Status)
}

type statuses struct {
	s map[int]Status
}

func New() Statuses {
	return &statuses{}
}

func (s *statuses) GetStatus(code int) Status {
	if st, ok := s.s[code]; ok {
		return st
	}
	return newStatus(code)
}

func (s *statuses) SetRawStatus(code int, m ...state.Manage) {
	if s.s == nil {
		s.s = make(map[int]Status)
	}
	s.s[code] = newStatus(code, m...)
}

func (s *statuses) SetStatus(st Status) {
	s.s[st.Code()] = st
}
