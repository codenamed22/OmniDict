package store

import (
	"strings"
	"sync"
	"time"

	pb_kv "omnidict/proto/kv"
)

type Lock struct {
	TxnID     string
	Held      bool
	Exclusive bool
}

type item struct {
	value      string
	expiration time.Time
}

type Store struct {
	mu     sync.RWMutex
	data   map[string]item
	locks  map[string]*Lock
	staged map[string][]*pb_kv.TxnOperation
}

func NewStore() *Store {
	return &Store{
		data:   make(map[string]item),
		locks:  make(map[string]*Lock),
		staged: make(map[string][]*pb_kv.TxnOperation),
	}
}

func (s *Store) Put(key, value string, ttl time.Duration) {
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
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	it, exists := s.data[key]
	if !exists {
		return "", false
	}

	if !it.expiration.IsZero() && time.Now().After(it.expiration) {
		return "", false
	}

	return it.value, true
}

func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	it, exists := s.data[key]
	if !exists {
		return false
	}
	if !it.expiration.IsZero() && time.Now().After(it.expiration) {
		return false
	}
	return true
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Store) GetAllKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k, it := range s.data {
		if it.expiration.IsZero() || time.Now().Before(it.expiration) {
			keys = append(keys, k)
		}
	}
	return keys
}

func (s *Store) GetKeysWithPrefix(prefix string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var keys []string
	for k, it := range s.data {
		if strings.HasPrefix(k, prefix) && (it.expiration.IsZero() || time.Now().Before(it.expiration)) {
			keys = append(keys, k)
		}
	}
	return keys
}

func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]item)
}

func (s *Store) GetTTL(key string) (time.Duration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	it, exists := s.data[key]
	if !exists {
		return 0, false
	}

	if it.expiration.IsZero() {
		return -1, true
	}

	remaining := time.Until(it.expiration)
	if remaining <= 0 {
		return 0, false
	}

	return remaining, true
}

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
	for key, it := range s.data {
		if !it.expiration.IsZero() && now.After(it.expiration) {
			delete(s.data, key)
		}
	}
}

func (s *Store) AcquireLock(key, txnID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock, exists := s.locks[key]
	if !exists {
		s.locks[key] = &Lock{TxnID: txnID, Held: true}
		return true
	}

	if lock.Held && lock.TxnID != txnID {
		return false
	}

	lock.TxnID = txnID
	lock.Held = true
	return true
}

func (s *Store) ReleaseLock(key, txnID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if lock, exists := s.locks[key]; exists && lock.TxnID == txnID {
		lock.Held = false
	}
}

func (s *Store) StageOperations(txnID string, ops []*pb_kv.TxnOperation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.staged[txnID] = ops
}

func (s *Store) CommitStagedOperations(txnID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ops, exists := s.staged[txnID]; exists {
		for _, op := range ops {
			switch op.Op {
			case pb_kv.TxnOperation_SET:
				s.Put(op.Key, op.Value, 0)
			case pb_kv.TxnOperation_DELETE:
				delete(s.data, op.Key)
			}
		}
		delete(s.staged, txnID)
	}
}

func (s *Store) ClearStagedOperations(txnID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.staged, txnID)
}
