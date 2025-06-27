package store

import (
	"errors"
	"sync"
	"time"
	"strings"
)

// Represents the key-value storage
type Store struct {
	mu       sync.RWMutex
	data     map[string]item
}

type item struct {
	value      string
	expiration time.Time
}

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrKeyExpired    = errors.New("key expired")
)

// Creates a new Store instance
func NewStore() *Store {
	return &Store{
		data: make(map[string]item),
	}
}

// Stores a key-value pair with optional TTL
func (s *Store) Put(key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	s.data[key] = item{
		value:      value,
		expiration: exp,
	}
	return nil
}

// Retrieves a value by key
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return "", false
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return "", false
	}

	return item.value, true
}

// Checks if key exists
func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	_, exists := s.data[key]
	if exists {
		// Check expiration
		item := s.data[key]
		if !item.expiration.IsZero() && time.Now().After(item.expiration) {
			return false
		}
	}
	return exists
}

// Removes a key-value pair
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// Returns all keys
func (s *Store) GetAllKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k, item := range s.data {
		if item.expiration.IsZero() || !time.Now().After(item.expiration) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Returns keys with prefix
func (s *Store) GetKeysWithPrefix(prefix string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var keys []string
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Clears all data
func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]item)
}

// Gets remaining TTL for a key
func (s *Store) GetTTL(key string) (time.Duration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return 0, false
	}

	if item.expiration.IsZero() {
		return -1, true // No expiration (permanent)
	}

	remaining := time.Until(item.expiration)
	if remaining <= 0 {
		return 0, false // Expired
	}

	return remaining, true
}

// Runs a background go routine to clean expired keys
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
	for key, item := range s.data {
		if !item.expiration.IsZero() && now.After(item.expiration) {
			delete(s.data, key)
		}
	}
}
