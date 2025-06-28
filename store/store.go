package store

import (
	"strings"
	"sync"
	"time"
)

// Store is an in‑memory thread‑safe key‑value store with optional TTL support.
type Store struct {
	mu   sync.RWMutex
	data map[string]item
}

type item struct {
	value      string
	expiration time.Time
}

// NewStore returns a fresh Store.
func NewStore() *Store {
	return &Store{data: make(map[string]item)}
}

// Put inserts/overwrites a key with an optional TTL (0 = no expiry).
func (s *Store) Put(key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	s.data[key] = item{value: value, expiration: exp}
	return nil
}

// Get returns the value and existence flag.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, ok := s.data[key]
	if !ok || (it.expiration != (time.Time{}) && time.Now().After(it.expiration)) {
		return "", false
	}
	return it.value, true
}

// Exists checks presence without returning value.
func (s *Store) Exists(key string) bool {
	_, ok := s.Get(key)
	return ok
}

// Delete removes a key.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// GetAllKeys returns every non‑expired key.
func (s *Store) GetAllKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	keys := make([]string, 0, len(s.data))
	for k, it := range s.data {
		if it.expiration.IsZero() || now.Before(it.expiration) {
			keys = append(keys, k)
		}
	}
	return keys
}

// GetKeysWithPrefix filters keys by prefix.
func (s *Store) GetKeysWithPrefix(prefix string) []string {
	all := s.GetAllKeys()
	var res []string
	for _, k := range all {
		if strings.HasPrefix(k, prefix) {
			res = append(res, k)
		}
	}
	return res
}

// Flush clears the store.
func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]item)
}

// GetTTL returns remaining TTL (‑1 no expiry, ‑2 absent).
func (s *Store) GetTTL(key string) (time.Duration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, ok := s.data[key]
	if !ok {
		return 0, false
	}
	if it.expiration.IsZero() {
		return -1, true
	}
	remaining := time.Until(it.expiration)
	if remaining < 0 {
		return 0, false
	}
	return remaining, true
}

// StartTTLCleaner launches a goroutine to delete expired keys.
func (s *Store) StartTTLCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.cleanupExpired()
		}
	}()
}

func (s *Store) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, it := range s.data {
		if !it.expiration.IsZero() && now.After(it.expiration) {
			delete(s.data, k)
		}
	}
}
