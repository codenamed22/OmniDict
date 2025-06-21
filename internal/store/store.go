package store

import "sync"

type KVStore struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewStore creates and returns a new in-memory store
func NewStore() *KVStore {
	return &KVStore{
		data: make(map[string]string),
	}
}

func (s *KVStore) Put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *KVStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}
