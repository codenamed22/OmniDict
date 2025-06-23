package store

import (
	// Package store implements a distributed key-value store using consistent hashing.
	"omnidict/internal/ring"
	"sync"
)

// Store represents the distributed key-value store system.
// It internally uses a consistent hash ring for sharding data across nodes.
type Store struct {
	hashRing *ring.HashRing // Consistent hash ring
	mutex    sync.RWMutex   // Thread safety
}

// NewStore creates and initializes a new distributed Store.
func NewStore() *Store {
	// Create a new hash ring with 3 virtual nodes per real node
	hr := ring.NewHashRing(3)

	// Add initial nodes (can be dynamic in production)
	hr.AddNode("node1")
	hr.AddNode("node2")
	hr.AddNode("node3")

	return &Store{
		hashRing: hr,
	}
}

// Put stores a key-value pair in the appropriate node using consistent hashing.
func (s *Store) Put(key, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Set the key-value pair on the correct node via hash ring
	s.hashRing.Set(key, value)
}

// Get retrieves the value for a key, looking up the correct node using hashing.
func (s *Store) Get(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.hashRing.Get(key)
}

// AddNode adds a new node to the hash ring.
func (s *Store) AddNode(nodeName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.hashRing.AddNode(nodeName)
}

// RemoveNode removes an existing node from the hash ring.
func (s *Store) RemoveNode(nodeName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.hashRing.RemoveNode(nodeName)
}

// PrintStatus prints the current state of the hash ring (for debugging/demo).
func (s *Store) PrintStatus() {
	s.hashRing.PrintRingStatus()
}
