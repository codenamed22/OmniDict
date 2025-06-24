package store

import (
	"sync"
	"time"
	"errors"
)

// represents the key-value storage
type Store struct {
	mu       sync.RWMutex
	data     map[string]item
	replicas map[string]map[string]item // replicaID -> key -> item
}

type item struct {
	value      []byte
	expiration time.Time
}

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrKeyExpired    = errors.New("key expired")
	ErrReplicaExists = errors.New("replica already exists")
)

// creates a new Store instance
func NewStore() *Store {
	return &Store{
		data:     make(map[string]item),
		replicas: make(map[string]map[string]item),
	}
}

// stores a key-value pair with optional TTL
func (s *Store) Put(key string, value []byte, ttl time.Duration) error {
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

// retrieves a value by key
func (s *Store) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return nil, ErrKeyExpired
	}

	return item.value, nil
}

// removes a key-value pair
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

//  handles data replication to other nodes
func (s *Store) Replicate(replicaID, key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.replicas[replicaID]; !exists {
		s.replicas[replicaID] = make(map[string]item)
	}

	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	s.replicas[replicaID][key] = item{
		value:      value,
		expiration: exp,
	}
	return nil
}

// retrieves replicated data
func (s *Store) GetReplica(replicaID, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	replica, exists := s.replicas[replicaID]
	if !exists {
		return nil, ErrKeyNotFound
	}

	item, exists := replica[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return nil, ErrKeyExpired
	}

	return item.value, nil
}

// runs a background go routine to clean expired keys
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

	// Clean main data
	for key, item := range s.data {
		if !item.expiration.IsZero() && now.After(item.expiration) {
			delete(s.data, key)
		}
	}

	// Clean replicas
	for replicaID, replica := range s.replicas {
		for key, item := range replica {
			if !item.expiration.IsZero() && now.After(item.expiration) {
				delete(replica, key)
			}
		}
		
		// Remove empty replicas
		if len(replica) == 0 {
			delete(s.replicas, replicaID)
		}
	}
}

// creates a consistent snapshot of the store
func (s *Store) Snapshot() map[string][]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := make(map[string][]byte)
	for key, item := range s.data {
		// Skip expired items
		if !item.expiration.IsZero() && time.Now().After(item.expiration) {
			continue
		}
		snapshot[key] = item.value
	}
	return snapshot
}

// loads data from a snapshot
func (s *Store) Restore(data map[string][]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, value := range data {
		s.data[key] = item{
			value:      value,
			expiration: time.Time{}, // No TTL for restored items
		}
	}
}

// returns storage statistics
func (s *Store) Stats() (keys int, size int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys = len(s.data)
	for _, item := range s.data {
		size += int64(len(item.value))
	}
	return keys, size
}
