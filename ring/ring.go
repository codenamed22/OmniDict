package ring

import (
	"context"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
	"time"

	pb_ring "omnidict/proto/ring"
	"omnidict/store"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HashRing implements consistent hashing for shard determination
type HashRing struct {
	virtualNodes int
	nodes        map[uint32]string // hash -> shardID
	sortedHashes []uint32
	mu           sync.RWMutex
}

func NewHashRing(virtualNodes int) *HashRing {
	return &HashRing{
		virtualNodes: virtualNodes,
		nodes:        make(map[uint32]string),
		sortedHashes: []uint32{},
	}
}

func (hr *HashRing) AddNode(shardID string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	for i := 0; i < hr.virtualNodes; i++ {
		virtualKey := shardID + ":" + strconv.Itoa(i)
		hash := hr.hash(virtualKey)
		hr.nodes[hash] = shardID
		hr.sortedHashes = append(hr.sortedHashes, hash)
	}
	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})
}

func (hr *HashRing) RemoveNode(shardID string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	newHashes := []uint32{}
	for hash, id := range hr.nodes {
		if id == shardID {
			delete(hr.nodes, hash)
		} else {
			newHashes = append(newHashes, hash)
		}
	}
	hr.sortedHashes = newHashes
	sort.Slice(hr.sortedHashes, func(i, j int) bool {
		return hr.sortedHashes[i] < hr.sortedHashes[j]
	})
}

func (hr *HashRing) GetShardID(key string) string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.sortedHashes) == 0 {
		return ""
	}

	hash := hr.hash(key)
	idx := sort.Search(len(hr.sortedHashes), func(i int) bool {
		return hr.sortedHashes[i] >= hash
	})

	if idx == len(hr.sortedHashes) {
		idx = 0
	}

	return hr.nodes[hr.sortedHashes[idx]]
}

func (hr *HashRing) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// RingServer handles gRPC requests using consistent hashing
type RingServer struct {
	pb_ring.UnimplementedRingServiceServer
	hashRing    *HashRing
	shardStores map[string]*store.Store // shardID -> store
	currentNode string
}

func NewRingServer(
	virtualNodes int,
	shardStores map[string]*store.Store,
	currentNode string,
) *RingServer {
	hashRing := NewHashRing(virtualNodes)
	return &RingServer{
		hashRing:    hashRing,
		shardStores: shardStores,
		currentNode: currentNode,
	}
}

func (s *RingServer) Put(ctx context.Context, req *pb_ring.PutRequest) (*pb_ring.PutResponse, error) {
	key := req.GetKey()
	shardID := s.hashRing.GetShardID(key)
	store, exists := s.shardStores[shardID]

	if !exists {
		return nil, status.Errorf(codes.NotFound, "shard %s not found", shardID)
	}

	// Local processing
	store.Put(key, string(req.GetValue()), time.Duration(req.Ttl)*time.Second)
	return &pb_ring.PutResponse{Success: true}, nil
}

func (s *RingServer) Get(ctx context.Context, req *pb_ring.GetRequest) (*pb_ring.GetResponse, error) {
	key := req.GetKey()
	shardID := s.hashRing.GetShardID(key)
	store, exists := s.shardStores[shardID]

	if !exists {
		return nil, status.Errorf(codes.NotFound, "shard %s not found", shardID)
	}

	value, exists := store.Get(key)
	return &pb_ring.GetResponse{
		Found: exists,
		Value: value,
	}, nil
}

func (s *RingServer) Delete(ctx context.Context, req *pb_ring.DeleteRequest) (*pb_ring.DeleteResponse, error) {
	key := req.GetKey()
	shardID := s.hashRing.GetShardID(key)
	store, exists := s.shardStores[shardID]

	if !exists {
		return nil, status.Errorf(codes.NotFound, "shard %s not found", shardID)
	}

	store.Delete(key)
	return &pb_ring.DeleteResponse{Success: true}, nil
}

func (s *RingServer) AddNode(ctx context.Context, req *pb_ring.AddNodeRequest) (*pb_ring.AddNodeResponse, error) {
	s.hashRing.AddNode(req.NodeId)
	return &pb_ring.AddNodeResponse{Success: true}, nil
}

func (s *RingServer) RemoveNode(ctx context.Context, req *pb_ring.RemoveNodeRequest) (*pb_ring.RemoveNodeResponse, error) {
	s.hashRing.RemoveNode(req.NodeId)
	return &pb_ring.RemoveNodeResponse{Success: true}, nil
}

func (s *RingServer) GetShard(ctx context.Context, req *pb_ring.GetShardRequest) (*pb_ring.GetShardResponse, error) {
	shardID := s.hashRing.GetShardID(req.Key)
	return &pb_ring.GetShardResponse{ShardId: shardID}, nil
}
