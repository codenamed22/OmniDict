package main

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strconv"
	"sync"
)

// consistent hash ring
type HashRing struct {
	nodes        map[uint32]string            // hash -> node name
	nodeNames    []string                     // list of actual nodes
	sortedHashes []uint32                     // sorted hash values for binary search
	virtualNodes int                          // number of virtual nodes per physical node
	nodeStores   map[string]map[string]string // node name -> its own key-value store

	// fields for cluster integration
	nodeAddresses map[string]string // node name -> gRPC address
	shardMapping  map[string]string // node name -> shard ID
	mutex         sync.RWMutex      // Thread safety for concurrent access
}

// creates a new consistent hash ring
func NewHashRing(virtualNodes int) *HashRing {
	return &HashRing{
		nodes:         make(map[uint32]string),
		nodeNames:     make([]string, 0),
		sortedHashes:  make([]uint32, 0),
		virtualNodes:  virtualNodes,
		nodeStores:    make(map[string]map[string]string),
		nodeAddresses: make(map[string]string),
		shardMapping:  make(map[string]string),
	}
}

// generates a hash value for a given key
func (hr *HashRing) hash(key string) uint32 {
	h := sha1.New()
	h.Write([]byte(key))
	hashBytes := h.Sum(nil)

	// Converting first 4 bytes to uint32
	return uint32(hashBytes[0])<<24 | uint32(hashBytes[1])<<16 |
		uint32(hashBytes[2])<<8 | uint32(hashBytes[3])
}

// adding a new node to the hash ring
func (hr *HashRing) AddNode(nodeName string) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	// Check if node already exists
	for _, name := range hr.nodeNames {
		if name == nodeName {
			fmt.Printf("Node %s already exists\n", nodeName)
			return
		}
	}

	hr.nodeNames = append(hr.nodeNames, nodeName)
	hr.nodeStores[nodeName] = make(map[string]string)

	// Adding virtual node : distribution
	for i := 0; i < hr.virtualNodes; i++ {
		virtualKey := nodeName + ":" + strconv.Itoa(i)
		hash := hr.hash(virtualKey)
		hr.nodes[hash] = nodeName
		hr.sortedHashes = append(hr.sortedHashes, hash)
	}

	// Hash sorting : efficient lookup
	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})

	fmt.Printf("Added node: %s with %d virtual nodes\n", nodeName, hr.virtualNodes)
}

// removes a node from the hash ring
func (hr *HashRing) RemoveNode(nodeName string) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	// Find and remove from nodeNames
	nodeIndex := -1
	for i, name := range hr.nodeNames {
		if name == nodeName {
			nodeIndex = i
			break
		}
	}

	if nodeIndex == -1 {
		fmt.Printf("Node %s not found\n", nodeName)
		return
	}

	// Remove from nodeNames slice
	hr.nodeNames = append(hr.nodeNames[:nodeIndex], hr.nodeNames[nodeIndex+1:]...)

	// Remove the node's store
	delete(hr.nodeStores, nodeName)

	// Remove virtual nodes
	newSortedHashes := make([]uint32, 0)
	for _, hash := range hr.sortedHashes {
		if hr.nodes[hash] != nodeName {
			newSortedHashes = append(newSortedHashes, hash)
		} else {
			delete(hr.nodes, hash)
		}
	}
	hr.sortedHashes = newSortedHashes

	fmt.Printf("Removed node: %s\n", nodeName)
}

// returns the node responsible for a given key
func (hr *HashRing) GetNode(key string) string {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	if len(hr.sortedHashes) == 0 {
		return ""
	}

	hash := hr.hash(key)

	// Find the first node hash that is >= key hash
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= hash
	})

	// If no node found, wrap around to first node
	if idx == len(hr.sortedHashes) {
		idx = 0
	}

	return hr.nodes[hr.sortedHashes[idx]]
}

// stores a key-value pair in specific node's store
func (hr *HashRing) Set(key, value string) {
	node := hr.GetNode(key)
	if node == "" {
		fmt.Printf("No nodes available to store key: %s\n", key)
		return
	}

	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	// Store in the specific node's store
	if nodeStore, exists := hr.nodeStores[node]; exists {
		nodeStore[key] = value
		fmt.Printf("Stored %s=%s on node %s\n", key, value, node)
	} else {
		fmt.Printf("Node %s store not found\n", node)
	}
}

// retrieves a value for a given key from the appropriate node's store
func (hr *HashRing) Get(key string) (string, bool) {
	node := hr.GetNode(key)
	if node == "" {
		return "", false
	}

	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	// Retrieve from the specific node's store
	if nodeStore, exists := hr.nodeStores[node]; exists {
		value, found := nodeStore[key]
		return value, found
	}

	return "", false
}

// returns all keys stored on a specific node
func (hr *HashRing) GetNodeKeys(nodeName string) []string {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	keys := make([]string, 0)

	if nodeStore, exists := hr.nodeStores[nodeName]; exists {
		for key := range nodeStore {
			keys = append(keys, key)
		}
	}

	return keys
}

// returns a copy of all key-value pairs for a specific node
func (hr *HashRing) GetNodeStore(nodeName string) map[string]string {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	result := make(map[string]string)

	if nodeStore, exists := hr.nodeStores[nodeName]; exists {
		for key, value := range nodeStore {
			result[key] = value
		}
	}

	return result
}

// returns the total number of keys across all nodes
func (hr *HashRing) GetTotalKeyCount() int {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	total := 0
	for _, nodeStore := range hr.nodeStores {
		total += len(nodeStore)
	}
	return total
}

// adds a new node with its gRPC address
func (hr *HashRing) AddNodeWithAddress(nodeName, address string) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	// Store the address mapping
	hr.nodeAddresses[nodeName] = address
	hr.AddNode(nodeName)
}

// returns the gRPC address for a node
func (hr *HashRing) GetNodeAddress(nodeName string) (string, bool) {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	address, exists := hr.nodeAddresses[nodeName]
	return address, exists
}

// returns keys that should move to a new node
func (hr *HashRing) GetKeysToMigrate(newNode string) []string {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	keysToMigrate := make([]string, 0)

	// Check all existing nodes for keys that should move to the new node
	for nodeName, nodeStore := range hr.nodeStores {
		if nodeName == newNode {
			continue // Skip the new node itself
		}

		for key := range nodeStore {
			// Check if this key now belongs to the new node
			currentResponsibleNode := hr.GetNode(key)
			if currentResponsibleNode == newNode {
				keysToMigrate = append(keysToMigrate, key)
			}
		}
	}

	return keysToMigrate
}

// moves a key from one node to another
func (hr *HashRing) MigrateKey(key, fromNode, toNode string) bool {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	// Get value from source node
	fromStore, fromExists := hr.nodeStores[fromNode]
	if !fromExists {
		return false
	}

	value, keyExists := fromStore[key]
	if !keyExists {
		return false
	}

	// Add to destination node
	toStore, toExists := hr.nodeStores[toNode]
	if !toExists {
		return false
	}

	toStore[key] = value
	delete(fromStore, key)

	fmt.Printf("Migrated key %s from %s to %s\n", key, fromNode, toNode)
	return true
}
func (hr *HashRing) PrintRingStatus() {
	fmt.Printf("\n=== Hash Ring Status ===\n")
	fmt.Printf("Nodes: %v\n", hr.nodeNames)
	fmt.Printf("Total virtual nodes: %d\n", len(hr.sortedHashes))
	fmt.Printf("Keys in store: %d\n", hr.GetTotalKeyCount())

	// Show key distribution per node
	for _, nodeName := range hr.nodeNames {
		keys := hr.GetNodeKeys(nodeName)
		fmt.Printf("Node %s: %d keys %v\n", nodeName, len(keys), keys)
	}
	fmt.Printf("-----------------------------\n\n")
}

// shows what happens when nodes are added/removed
func (hr *HashRing) SimulateKeyMovement(keys []string) {
	fmt.Printf("=== Simulating Key Movement ===\n")

	// Store initial mapping
	fmt.Printf("Initial key distribution:\n")
	initialMapping := make(map[string]string)
	for _, key := range keys {
		node := hr.GetNode(key)
		initialMapping[key] = node
		fmt.Printf("  %s -> %s\n", key, node)
	}

	// Add a new node
	fmt.Printf("\nAdding new node 'node-new'...\n")
	hr.AddNode("node-new")

	// Check new mapping
	fmt.Printf("Key distribution after adding node:\n")
	moved := 0
	for _, key := range keys {
		newNode := hr.GetNode(key)
		oldNode := initialMapping[key]
		if newNode != oldNode {
			fmt.Printf("  %s: %s -> %s (MOVED)\n", key, oldNode, newNode)
			moved++
		} else {
			fmt.Printf("  %s -> %s\n", key, newNode)
		}
	}

	fmt.Printf("Keys moved: %d/%d (%.1f%%)\n", moved, len(keys), float64(moved)/float64(len(keys))*100)
	fmt.Printf("===============================\n\n")
}
