package ring

import (
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

// ConsistentHashRing represents the consistent hashing ring
type ConsistentHashRing struct {
	mu          sync.RWMutex
	replicas    int
	keys        []int              // Sorted
	hashMap     map[int]string     // hash → node
	nodeSet     map[string]struct{} // For tracking unique nodes
}

// NewConsistentHashRing creates a ring with default 100 replicas
func NewConsistentHashRing() *ConsistentHashRing {
	return &ConsistentHashRing{
		replicas: 100,
		hashMap:  make(map[int]string),
		nodeSet:  make(map[string]struct{}),
	}
}

// AddNode adds a node to the ring
func (r *ConsistentHashRing) AddNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock(
