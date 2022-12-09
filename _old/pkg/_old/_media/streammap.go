package media

import (
	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type StreamMap struct {
	m map[int]*Stream
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewStreamMap() *StreamMap {
	m := new(StreamMap)
	m.m = make(map[int]*Stream)
	return m
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (m *StreamMap) Set(s *Stream) error {
	k := s.Index()
	if _, exists := m.m[k]; exists {
		return ErrDuplicateEntry.With(s)
	}
	m.m[k] = s
	return nil
}

func (m *StreamMap) Get(k int) *Stream {
	return m.m[k]
}
