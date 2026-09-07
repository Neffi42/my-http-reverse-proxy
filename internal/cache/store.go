package cache

import (
	"net/http"
	"sync"
	"time"
)

type Entry struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	StoredAt   time.Time
	ExpiresAt  time.Time
}

func (e *Entry) Fresh(now time.Time) bool {
	return now.Before(e.ExpiresAt)
}

func (e *Entry) Age(now time.Time) time.Duration {
	return now.Sub(e.StoredAt)
}

type Store struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

func NewStore() *Store {
	return &Store{entries: make(map[string]*Entry)}
}

func (s *Store) Get(key string) (*Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[key]
	return e, ok
}

func (s *Store) Set(key string, e *Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = e
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}
