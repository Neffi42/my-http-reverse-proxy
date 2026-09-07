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
