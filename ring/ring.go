package ring

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"encoding/json"
	"errors"
	"time"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	pb "omnidict/proto"
	"omnidict/store"

)

// TTL support
type keyMetadata struct {
	value 	string
	expiration time.Time
}

// implements the gRPC server for the consistent hash ring
type RingServer struct {
    pb.UnimplementedRingServiceServer   // Embedded for forward compatibility
    hashRing  *consistenthash.Ring      // Your consistent hashing implementation
    store     *store.Store              // Connection to storage module
    currentNode string                  // Current node identifier
	grpcServer *grpc.Server             // gRPC server instance
}

// consistent hash ring
type HashRing struct {
	mu 			 sync.RWMutex 				  // Mutex for thread safety
	nodes        map[uint32]string            // hash -> node name
	nodeNames    []string                     // list of actual nodes
	sortedHashes []uint32                     // sorted hash values for binary search
	virtualNodes int                          // number of virtual nodes per physical node
	nodeStores   map[string]map[string]string // node name -> its own key-value store

	// TTL support
	keyMetadata map[string]keyMetadata        // key -> {value, expiration}

	// fields for cluster integration
	nodeAddresses map[string]string // node name -> gRPC address
	shardMapping  map[string]string // node name -> shard ID
	
	// New fields for enhanced functionality
	nodeHealth    map[string]bool      // node name -> health status
	nodeLoad      map[string]int       // node name -> current load (key count)
	replicationFactor int              // number of replicas for each key
	lastHealthCheck   time.Time        // last time health check was performed
	consistencyLevel  ConsistencyLevel // consistency level for operations
}

// ConsistencyLevel defines different consistency guarantees
type ConsistencyLevel int

const (
	ONE ConsistencyLevel = iota // Wait for one node to respond
	QUORUM                      // Wait for majority of replicas
	ALL                         // Wait for all replicas
)

// NodeInfo contains information about a node
type NodeInfo struct {
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Healthy   bool      `json:"healthy"`
	KeyCount  int       `json:"key_count"`
	LastSeen  time.Time `json:"last_seen"`
}

// ReplicationInfo contains replication details for a key
type ReplicationInfo struct {
	PrimaryNode string   `json:"primary_node"`
	Replicas    []string `json:"replicas"`
}

// creates a new consistent hash ring with replication support
func NewHashRingWithReplication(virtualNodes, replicationFactor int) *HashRing {
	return &HashRing{
		nodes:             make(map[uint32]string),
		nodeNames:         make([]string, 0),
		sortedHashes:      make([]uint32, 0),
		virtualNodes:      virtualNodes,
		nodeStores:        make(map[string]map[string]string),
		nodeAddresses:     make(map[string]string),
		shardMapping:      make(map[string]string),
		nodeHealth:        make(map[string]bool),
		nodeLoad:          make(map[string]int),
		replicationFactor: replicationFactor,
		consistencyLevel:  QUORUM,
		lastHealthCheck:   time.Now(),
		keyMetadata:       make(map[string]keyMetadata),
	}
}

// creates a new consistent hash ring
func NewHashRing(virtualNodes int) *HashRing {
	return NewHashRingWithReplication(virtualNodes, 1)
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
	hr.nodeHealth[nodeName] = true
	hr.nodeLoad[nodeName] = 0

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

	// Remove the node's store and metadata
	delete(hr.nodeStores, nodeName)
	delete(hr.nodeHealth, nodeName)
	delete(hr.nodeLoad, nodeName)
	delete(hr.nodeAddresses, nodeName)

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

// NEW: Returns N healthy nodes responsible for a key (for replication)
func (hr *HashRing) GetNodesForKey(key string, count int) []string {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	if len(hr.sortedHashes) == 0 {
		return []string{}
	}

	hash := hr.hash(key)
	nodes := make([]string, 0, count)
	seen := make(map[string]bool)

	// Find starting position
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= hash
	})

	// Collect unique healthy nodes
	for len(nodes) < count && len(seen) < len(hr.nodeNames) {
		if idx >= len(hr.sortedHashes) {
			idx = 0
		}

		nodeName := hr.nodes[hr.sortedHashes[idx]]
		if !seen[nodeName] && hr.nodeHealth[nodeName] {
			nodes = append(nodes, nodeName)
			seen[nodeName] = true
		}
		idx++
	}

	return nodes
}

// NEW: Get replication info for a key
func (hr *HashRing) GetReplicationInfo(key string) ReplicationInfo {
	nodes := hr.GetNodesForKey(key, hr.replicationFactor)
	if len(nodes) == 0 {
		return ReplicationInfo{}
	}

	return ReplicationInfo{
		PrimaryNode: nodes[0],
		Replicas:    nodes,
	}
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
		hr.nodeLoad[node] = len(nodeStore)
		fmt.Printf("Stored %s=%s on node %s\n", key, value, node)
	} else {
		fmt.Printf("Node %s store not found\n", node)
	}
}

// NEW: Set with replication support
func (hr *HashRing) SetWithReplication(key, value string, ttl time.Duration) error {
	nodes := hr.GetNodesForKey(key, hr.replicationFactor)
	if len(nodes) == 0 {
		return errors.New("no healthy nodes available")
	}

	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	metadata := keyMetadata{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	hr.keyMetadata[key] = metadata
	
	successCount := 0
	for _, nodeName := range nodes {
		if nodeStore, exists := hr.nodeStores[nodeName]; exists && hr.nodeHealth[nodeName] {
			nodeStore[key] = value
			hr.nodeLoad[nodeName] = len(nodeStore)
			successCount++
			fmt.Printf("Stored %s=%s on replica node %s\n", key, value, nodeName)
		}
	}

	// Check if we met consistency requirements
	required := hr.getRequiredSuccessCount(len(nodes))
	if successCount < required {
		return fmt.Errorf("insufficient replicas: got %d, need %d", successCount, required)
	}

	return nil
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

// NEW: Get with replication support (read repair)
func (hr *HashRing) GetWithReplication(key string) (string, error) {

	// expiration checking
	hr.mutex.RLock()
	metadata, exists := hr.keyMetadata[key]
	hr.mutex.RUnlock()

	if !exists {
		return "", errors.New("key not found")
	}

	if time.Now().After(metadata.expiration) {
		hr.Delete(key) // delete expired key
		return "", errors.New("key expired")
	}

	nodes := hr.GetNodesForKey(key, hr.replicationFactor)
	if len(nodes) == 0 {
		return "", errors.New("no healthy nodes available")
	}

	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	values := make(map[string]int) // value -> count
	var mostCommonValue string
	maxCount := 0

	for _, nodeName := range nodes {
		if nodeStore, exists := hr.nodeStores[nodeName]; exists && hr.nodeHealth[nodeName] {
			if value, found := nodeStore[key]; found {
				values[value]++
				if values[value] > maxCount {
					maxCount = values[value]
					mostCommonValue = value
				}
			}
		}
	}

	required := hr.getRequiredSuccessCount(len(nodes))
	if maxCount < required {
		return "", errors.New("insufficient replicas responded")
	}

	return mostCommonValue, nil
}

// expiration check helper method
func (hr *HashRing) isKeyExpired(key string) bool {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()
	metadata, exists := hr.keyMetadata[key]
	return exists && time.Now().After(metadata.expiration)
}

// TTL
func (hr *HashRing) GetTTL(key string) (time.Duration, bool) {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	metadata, exists := hr.keyMetadata[key]
	if !exists {
		return 0, false
	}

	if time.Now().After(metadata.expiration) {
		return 0, false // key expired
	}

	return time.Until(metadata.expiration), true
}

// NEW: Delete a key from all replicas
func (hr *HashRing) Delete(key string) error {
	nodes := hr.GetNodesForKey(key, hr.replicationFactor)
	if len(nodes) == 0 {
		return errors.New("no healthy nodes available")
	}

	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	successCount := 0
	for _, nodeName := range nodes {
		if nodeStore, exists := hr.nodeStores[nodeName]; exists && hr.nodeHealth[nodeName] {
			if _, found := nodeStore[key]; found {
				delete(nodeStore, key)
				hr.nodeLoad[nodeName] = len(nodeStore)
				successCount++
				fmt.Printf("Deleted %s from node %s\n", key, nodeName)
			}
		}
	}

	required := hr.getRequiredSuccessCount(len(nodes))
	if successCount < required {
		return fmt.Errorf("insufficient replicas: got %d, need %d", successCount, required)
	}

	return nil
}

// NEW: Check if a key exists
func (hr *HashRing) Exists(key string) bool {
	_, exists := hr.Get(key)
	return exists
}

// NEW: Get all keys across all nodes
func (hr *HashRing) GetAllKeys() []string {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	keySet := make(map[string]bool)
	for _, nodeStore := range hr.nodeStores {
		for key := range nodeStore {
			keySet[key] = true
		}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}

	return keys
}

// NEW: Get keys with a prefix
func (hr *HashRing) GetKeysWithPrefix(prefix string) []string {
	allKeys := hr.GetAllKeys()
	result := make([]string, 0)

	for _, key := range allKeys {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			result = append(result, key)
		}
	}

	return result
}

// NEW: Health check functionality
func (hr *HashRing) MarkNodeUnhealthy(nodeName string) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	if _, exists := hr.nodeHealth[nodeName]; exists {
		hr.nodeHealth[nodeName] = false
		fmt.Printf("Marked node %s as unhealthy\n", nodeName)
	}
}

func (hr *HashRing) MarkNodeHealthy(nodeName string) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	if _, exists := hr.nodeHealth[nodeName]; exists {
		hr.nodeHealth[nodeName] = true
		fmt.Printf("Marked node %s as healthy\n", nodeName)
	}
}

func (hr *HashRing) IsNodeHealthy(nodeName string) bool {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	healthy, exists := hr.nodeHealth[nodeName]
	return exists && healthy
}

// NEW: Get healthy nodes only
func (hr *HashRing) GetHealthyNodes() []string {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	healthy := make([]string, 0)
	for nodeName, isHealthy := range hr.nodeHealth {
		if isHealthy {
			healthy = append(healthy, nodeName)
		}
	}

	return healthy
}

// NEW: Load balancing - get least loaded node
func (hr *HashRing) GetLeastLoadedNode() string {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	var leastLoaded string
	minLoad := int(^uint(0) >> 1) // max int

	for nodeName, load := range hr.nodeLoad {
		if hr.nodeHealth[nodeName] && load < minLoad {
			minLoad = load
			leastLoaded = nodeName
		}
	}

	return leastLoaded
}

// NEW: Get node information
func (hr *HashRing) GetNodeInfo(nodeName string) (NodeInfo, error) {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	if _, exists := hr.nodeHealth[nodeName]; !exists {
		return NodeInfo{}, errors.New("node not found")
	}

	return NodeInfo{
		Name:     nodeName,
		Address:  hr.nodeAddresses[nodeName],
		Healthy:  hr.nodeHealth[nodeName],
		KeyCount: hr.nodeLoad[nodeName],
		LastSeen: time.Now(),
	}, nil
}

// NEW: Get cluster information
func (hr *HashRing) GetClusterInfo() map[string]NodeInfo {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	info := make(map[string]NodeInfo)
	for nodeName := range hr.nodeHealth {
		info[nodeName] = NodeInfo{
			Name:     nodeName,
			Address:  hr.nodeAddresses[nodeName],
			Healthy:  hr.nodeHealth[nodeName],
			KeyCount: hr.nodeLoad[nodeName],
			LastSeen: time.Now(),
		}
	}

	return info
}

// NEW: Set consistency level
func (hr *HashRing) SetConsistencyLevel(level ConsistencyLevel) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()
	hr.consistencyLevel = level
}

// NEW: Serialize hash ring state
func (hr *HashRing) SerializeState() ([]byte, error) {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	state := map[string]interface{}{
		"nodes":              hr.nodeNames,
		"node_addresses":     hr.nodeAddresses,
		"node_health":        hr.nodeHealth,
		"node_load":          hr.nodeLoad,
		"virtual_nodes":      hr.virtualNodes,
		"replication_factor": hr.replicationFactor,
		"consistency_level":  hr.consistencyLevel,
	}

	return json.Marshal(state)
}

// Helper function to determine required success count based on consistency level
func (hr *HashRing) getRequiredSuccessCount(totalReplicas int) int {
	switch hr.consistencyLevel {
	case ONE:
		return 1
	case QUORUM:
		return (totalReplicas / 2) + 1
	case ALL:
		return totalReplicas
	default:
		return 1
	}
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
	hr.nodeAddresses[nodeName] = address
	hr.mutex.Unlock()
	
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

	// Update load counters
	hr.nodeLoad[fromNode] = len(fromStore)
	hr.nodeLoad[toNode] = len(toStore)

	fmt.Printf("Migrated key %s from %s to %s\n", key, fromNode, toNode)
	return true
}

// NEW: Bulk migration of keys
func (hr *HashRing) BulkMigrateKeys(keys []string, fromNode, toNode string) int {
	successCount := 0
	for _, key := range keys {
		if hr.MigrateKey(key, fromNode, toNode) {
			successCount++
		}
	}
	return successCount
}

// Enhanced print function with health and load info
func (hr *HashRing) PrintRingStatus() {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	fmt.Printf("\n=== Hash Ring Status ===\n")
	fmt.Printf("Nodes: %v\n", hr.nodeNames)
	fmt.Printf("Total virtual nodes: %d\n", len(hr.sortedHashes))
	fmt.Printf("Keys in store: %d\n", hr.GetTotalKeyCount())
	fmt.Printf("Replication factor: %d\n", hr.replicationFactor)
	fmt.Printf("Consistency level: %v\n", hr.consistencyLevel)

	// Show detailed node information
	for _, nodeName := range hr.nodeNames {
		keys := hr.GetNodeKeys(nodeName)
		health := "HEALTHY"
		if !hr.nodeHealth[nodeName] {
			health = "UNHEALTHY"
		}
		fmt.Printf("Node %s [%s]: %d keys %v\n", nodeName, health, len(keys), keys)
		if addr, exists := hr.nodeAddresses[nodeName]; exists {
			fmt.Printf("  Address: %s\n", addr)
		}
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

func (hr *HashRing) GetNodeNames() []string {
    hr.mutex.RLock()
    defer hr.mutex.RUnlock()
    return hr.nodeNames
}

//gRPC Serve Initialization
func NewRingServer(virtualNodes, replicationFactor int, store *store.Store, port string) *RingServer {
	currentNode := fmt.Sprintf("localhost:%s", port)
	hashRing := NewHashRingWithReplication(virtualNodes, replicationFactor)
	hashRing.AddNodeWithAddress(currentNode, currentNode)
	
	return &RingServer{
		hashRing:   hashRing,
		store:      store,
		currentNode: currentNode,
	}
}

func (s *RingServer) Start() error {
	lis, err := net.Listen("tcp", s.currentNode)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.grpcServer = grpc.NewServer()
	pb.RegisterRingServiceServer(s.grpcServer, s)
	
	log.Printf("Server started at %s", s.currentNode)
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func (s *RingServer) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// gRPC Server Implemnentation
func (s *RingServer) GetNodeForRequest(ctx context.Context, req *pb.NodeRequest) (*pb.NodeResponse, error) {
	targetNode := s.hashRing.GetNode(req.Key)
	return &pb.NodeResponse{Node: targetNode}, nil
}

func (s *RingServer) AddNodeForRequest(ctx context.Context, req *pb.NodeRequest) (*pb.Empty, error) {
	s.hashRing.AddNodeWithAddress(req.Node, req.Address)
	return &pb.Empty{}, nil
}

func (s *RingServer) RemoveNodeForRequest(ctx context.Context, req *pb.NodeRequest) (*pb.Empty, error) {
	s.hashRing.RemoveNode(req.Node)
	return &pb.Empty{}, nil
}

// gRPC methods for Put, Get, Delete, and HealthCheck - add others
func (s *RingServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	key := req.GetKey()
	targetNode := s.hashRing.GetNode(key)

	if targetNode == s.currentNode {
		if err := s.store.Put(key, req.GetValue(), time.Duration(req.Ttl)*time.Second); err != nil {
			return nil, status.Errorf(codes.Internal, "storage error: %v", err)
		}
		return &pb.PutResponse{Success: true}, nil
	}
	return s.forwardPutRequest(targetNode, req)
}

func (s *RingServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	key := req.GetKey()
	targetNode := s.hashRing.GetNode(key)

	if targetNode == s.currentNode {
		value, err := s.store.Get(key)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "key not found: %v", err)
		}
		return &pb.GetResponse{Value: value}, nil
	}
	return s.forwardGetRequest(targetNode, req)
}

func (s *RingServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	key := req.GetKey()
	targetNode := s.hashRing.GetNode(key)

	if targetNode == s.currentNode {
		s.store.Delete(key)
		return &pb.DeleteResponse{Success: true}, nil
	}
	return s.forwardDeleteRequest(targetNode, req)
}


func (s *RingServer) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	status := s.hashRing.IsNodeHealthy(req.Node)
	return &pb.HealthResponse{Healthy: status}, nil
}

// gRPC Forwarding Logic - add other cmds as needed
func (s *RingServer) forwardPutRequest(node string, req *pb.PutRequest) (*pb.PutResponse, error) {
	conn, err := grpc.Dial(node, grpc.WithInsecure())
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "forwarding failed: %v", err)
	}
	defer conn.Close()
	client := pb.NewRingServiceClient(conn)
	return client.Put(context.Background(), req)
}

func (s *RingServer) forwardGetRequest(node string, req *pb.GetRequest) (*pb.GetResponse, error) {
	conn, err := grpc.Dial(node, grpc.WithInsecure())
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "forwarding failed: %v", err)
	}
	defer conn.Close()
	client := pb.NewRingServiceClient(conn)
	return client.Get(context.Background(), req)
}

func (s *RingServer) forwardDeleteRequest(node string, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	conn, err := grpc.Dial(node, grpc.WithInsecure())
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "forwarding failed: %v", err)
	}
	defer conn.Close()
	client := pb.NewRingServiceClient(conn)
	return client.Delete(context.Background(), req)
}