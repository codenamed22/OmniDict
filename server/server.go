package server

import (
	"context"
	"time"
	"fmt"
	"strings"

	pb_kv "omnidict/proto/kv"
	// pb_ring "omnidict/proto/ring" will be used later
	"omnidict/ring"
	"omnidict/store"

)

type OmnidictServer struct {
	pb_kv.UnimplementedOmnidictServiceServer
	Ring *ring.HashRing
	Storage *store.Store
}

func NewOmnidictServer() *OmnidictServer {
	return &OmnidictServer{
		Ring: ring.NewHashRingWithReplication(3, 3),
	}
}

func NewOmnidictServerWithStorage(storage *store.Store) *OmnidictServer {
	// hashring creation and initialization with storage-backed nodes
	hashRing := ring.NewHashRing(3)

	// nodes with storage backends
	for i := 0; i < 3; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		hashRing.AddNode(nodeID)
	}

	return &OmnidictServer{
		Ring: ring.NewHashRingWithReplication(3, 3),
		Storage: storage,
	}
}

func (s *OmnidictServer) Put(ctx context.Context, req *pb_kv.PutRequest) (*pb_kv.PutResponse, error) {
	// Use storage directly if ring has issues
	if s.Storage != nil {
		err := s.Storage.Put(req.Key, req.Value, 0) // 0 TTL = no expiration
		if err != nil {
			return &pb_kv.PutResponse{Success: false, Message: err.Error()}, nil
		}
		return &pb_kv.PutResponse{Success: true, Message: "OK"}, nil
	}
	
	// Fallback to ring if storage not available
	err := s.Ring.PutWithReplication(req.Key, req.Value, 0)
	if err != nil {
		return &pb_kv.PutResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb_kv.PutResponse{Success: true, Message: "OK"}, nil
}

func (s *OmnidictServer) Get(ctx context.Context, req *pb_kv.GetRequest) (*pb_kv.GetResponse, error) {
	// Use storage directly
	if s.Storage != nil {
		value, exists := s.Storage.Get(req.Key)
		if !exists {
			return &pb_kv.GetResponse{Found: false, Value: ""}, nil
		}
		return &pb_kv.GetResponse{Found: true, Value: value}, nil
	}
	
	// Fallback to ring
	value, err := s.Ring.GetWithReplication(req.Key)
	if err != nil {
		return &pb_kv.GetResponse{Found: false, Value: ""}, nil
	}
	return &pb_kv.GetResponse{Found: true, Value: value}, nil
}

func (s *OmnidictServer) Delete(ctx context.Context, req *pb_kv.DeleteRequest) (*pb_kv.DeleteResponse, error) {
	if s.Storage != nil {
		s.Storage.Delete(req.Key)
		return &pb_kv.DeleteResponse{Success: true}, nil
	}
	
	// Fallback to ring
	err := s.Ring.Delete(req.Key)
	return &pb_kv.DeleteResponse{Success: err == nil}, err
}

func (s *OmnidictServer) Exists(ctx context.Context, req *pb_kv.ExistsRequest) (*pb_kv.ExistsResponse, error) {
	if s.Storage != nil {
		_, exists := s.Storage.Get(req.Key)
		return &pb_kv.ExistsResponse{Exists: exists}, nil
	}
	
	// Fallback to ring
	exists := s.Ring.Exists(req.Key)
	return &pb_kv.ExistsResponse{Exists: exists}, nil
}

func (s *OmnidictServer) Keys(ctx context.Context, req *pb_kv.KeysRequest) (*pb_kv.KeysResponse, error) {
	if s.Storage != nil {
		keys := s.Storage.GetAllKeys()
		if req.Pattern != "" {
			filtered := []string{}
			for _, key := range keys {
				if strings.HasPrefix(key, req.Pattern) {
					filtered = append(filtered, key)
				}
			}
			keys = filtered
		}
		return &pb_kv.KeysResponse{Keys: keys}, nil
	}
	
	// Fallback to ring
	var keys []string
	if req.Pattern == "" {
		keys = s.Ring.GetAllKeys()
	} else {
		keys = s.Ring.GetKeysWithPrefix(req.Pattern)
	}
	return &pb_kv.KeysResponse{Keys: keys}, nil
}

func (s *OmnidictServer) Flush(ctx context.Context, req *pb_kv.FlushRequest) (*pb_kv.FlushResponse, error) {
	if s.Storage != nil {
		s.Storage.Flush()
		return &pb_kv.FlushResponse{Success: true}, nil
	}
	
	// Fallback to ring
	for _, node := range s.Ring.GetNodeNames() {
		store := s.Ring.GetNodeStore(node)
		for key := range store {
			s.Ring.Delete(key)
		}
	}
	return &pb_kv.FlushResponse{Success: true}, nil
}

func (s *OmnidictServer) Expire(ctx context.Context, req *pb_kv.ExpireRequest) (*pb_kv.ExpireResponse, error) {
	if s.Storage != nil {
		// Get current value and Put with TTL
		if val, exists := s.Storage.Get(req.Key); exists {
			s.Storage.Put(req.Key, val, time.Duration(req.Ttl)*time.Second)
			return &pb_kv.ExpireResponse{Success: true, Message: "Expiration Put"}, nil
		}
		return &pb_kv.ExpireResponse{Success: false, Message: "Key not found"}, nil
	}
	
	// Fallback to ring
	value, err := s.Ring.GetWithReplication(req.Key)
	if err != nil {
		return &pb_kv.ExpireResponse{Success: false, Message: "Key not found"}, nil
	}
	err = s.Ring.PutWithReplication(req.Key, value, time.Duration(req.Ttl)*time.Second)
	if err != nil {
		return &pb_kv.ExpireResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb_kv.ExpireResponse{Success: true, Message: "Expiration Put"}, nil
}

func (s *OmnidictServer) TTL(ctx context.Context, req *pb_kv.TTLRequest) (*pb_kv.TTLResponse, error) {
	if s.Storage != nil {
		ttl, exists := s.Storage.GetTTL(req.Key)
		if !exists {
			return &pb_kv.TTLResponse{Ttl: -2}, nil // Key doesn't exist
		}
		return &pb_kv.TTLResponse{Ttl: int64(ttl.Seconds())}, nil
	}
	
	// Fallback to ring
	ttl, exists := s.Ring.GetTTL(req.Key)
	if !exists {
		return &pb_kv.TTLResponse{Ttl: -2}, nil
	}
	return &pb_kv.TTLResponse{Ttl: int64(ttl.Seconds())}, nil
}

func (s *OmnidictServer) Update(ctx context.Context, req *pb_kv.UpdateRequest) (*pb_kv.UpdateResponse, error) {
	if s.Storage != nil {
		if _, exists := s.Storage.Get(req.Key); !exists {
			return &pb_kv.UpdateResponse{Success: false, Message: "Key not found"}, nil
		}
		
		// Preserve existing TTL
		ttl, _ := s.Storage.GetTTL(req.Key)
		s.Storage.Put(req.Key, req.Value, ttl)
		return &pb_kv.UpdateResponse{Success: true, Message: "Updated successfully"}, nil
	}
	
	// Fallback to ring
	if !s.Ring.Exists(req.Key) {
		return &pb_kv.UpdateResponse{Success: false, Message: "Key not found"}, nil
	}
	ttl, _ := s.Ring.GetTTL(req.Key)
	err := s.Ring.PutWithReplication(req.Key, req.Value, ttl)
	if err != nil {
		return &pb_kv.UpdateResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb_kv.UpdateResponse{Success: true, Message: "Updated successfully"}, nil
}

func (s *OmnidictServer) GetNodeInfo(ctx context.Context, req *pb_kv.NodeInfoRequest) (*pb_kv.NodeInfoResponse, error) {
	// Node info remains ring-specific
	nodeNames := s.Ring.GetNodeNames()
	return &pb_kv.NodeInfoResponse{
		NodeId:     "node-1",
		Address:    "localhost:50051",
		Status:     "healthy",
		TotalNodes: int32(len(nodeNames)),
		Nodes:      nodeNames,
	}, nil
}
