package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb_kv "omnidict/proto/kv"
	"omnidict/ring"
	"omnidict/store"

	"github.com/hashicorp/raft"
)

type OmnidictServer struct {
	pb_kv.UnimplementedOmnidictServiceServer
	Ring    *ring.HashRing
	Storage *store.Store
	Raft    *raft.Raft
}

func NewOmnidictServerWithRaft(storage *store.Store, r *raft.Raft) *OmnidictServer {
	hashRing := ring.NewHashRing(3)
	for i := 0; i < 3; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		hashRing.AddNode(nodeID)
	}

	return &OmnidictServer{
		Ring:    ring.NewHashRingWithReplication(3, 3),
		Storage: storage,
		Raft:    r,
	}
}

type Command struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

func (s *OmnidictServer) Put(ctx context.Context, req *pb_kv.PutRequest) (*pb_kv.PutResponse, error) {
	if s.Raft != nil && s.Raft.State() == raft.Leader {
		cmd := Command{Op: "put", Key: req.Key, Value: req.Value}
		data, _ := json.Marshal(cmd)
		f := s.Raft.Apply(data, 5*time.Second)
		if err := f.Error(); err != nil {
			return &pb_kv.PutResponse{Success: false, Message: err.Error()}, nil
		}
		return &pb_kv.PutResponse{Success: true, Message: "Put via Raft"}, nil
	}

	if s.Storage != nil {
		err := s.Storage.Put(req.Key, req.Value, 0)
		if err != nil {
			return &pb_kv.PutResponse{Success: false, Message: err.Error()}, nil
		}
		return &pb_kv.PutResponse{Success: true, Message: "Put in store"}, nil
	}
	err := s.Ring.PutWithReplication(req.Key, req.Value, 0)
	return &pb_kv.PutResponse{Success: err == nil, Message: "Put in ring"}, nil
}

func (s *OmnidictServer) Delete(ctx context.Context, req *pb_kv.DeleteRequest) (*pb_kv.DeleteResponse, error) {
	if s.Raft != nil && s.Raft.State() == raft.Leader {
		cmd := Command{Op: "delete", Key: req.Key}
		data, _ := json.Marshal(cmd)
		f := s.Raft.Apply(data, 5*time.Second)
		if err := f.Error(); err != nil {
			return &pb_kv.DeleteResponse{Success: false}, nil
		}
		return &pb_kv.DeleteResponse{Success: true}, nil
	}

	if s.Storage != nil {
		s.Storage.Delete(req.Key)
		return &pb_kv.DeleteResponse{Success: true}, nil
	}
	err := s.Ring.Delete(req.Key)
	return &pb_kv.DeleteResponse{Success: err == nil}, err
}

func (s *OmnidictServer) Update(ctx context.Context, req *pb_kv.UpdateRequest) (*pb_kv.UpdateResponse, error) {
	if s.Raft != nil && s.Raft.State() == raft.Leader {
		cmd := Command{Op: "put", Key: req.Key, Value: req.Value}
		data, _ := json.Marshal(cmd)
		f := s.Raft.Apply(data, 5*time.Second)
		if err := f.Error(); err != nil {
			return &pb_kv.UpdateResponse{Success: false, Message: err.Error()}, nil
		}
		return &pb_kv.UpdateResponse{Success: true, Message: "Updated via Raft"}, nil
	}

	if s.Storage != nil {
		if _, exists := s.Storage.Get(req.Key); !exists {
			return &pb_kv.UpdateResponse{Success: false, Message: "Key not found"}, nil
		}
		ttl, _ := s.Storage.GetTTL(req.Key)
		s.Storage.Put(req.Key, req.Value, ttl)
		return &pb_kv.UpdateResponse{Success: true, Message: "Updated in store"}, nil
	}

	if !s.Ring.Exists(req.Key) {
		return &pb_kv.UpdateResponse{Success: false, Message: "Key not found"}, nil
	}
	ttl, _ := s.Ring.GetTTL(req.Key)
	err := s.Ring.PutWithReplication(req.Key, req.Value, ttl)
	return &pb_kv.UpdateResponse{Success: err == nil, Message: "Updated in ring"}, nil
}

func (s *OmnidictServer) Get(ctx context.Context, req *pb_kv.GetRequest) (*pb_kv.GetResponse, error) {
	if s.Storage != nil {
		value, exists := s.Storage.Get(req.Key)
		return &pb_kv.GetResponse{Found: exists, Value: value}, nil
	}
	value, err := s.Ring.GetWithReplication(req.Key)
	if err != nil {
		return &pb_kv.GetResponse{Found: false, Value: ""}, nil
	}
	return &pb_kv.GetResponse{Found: true, Value: value}, nil
}

func (s *OmnidictServer) Exists(ctx context.Context, req *pb_kv.ExistsRequest) (*pb_kv.ExistsResponse, error) {
	if s.Storage != nil {
		_, exists := s.Storage.Get(req.Key)
		return &pb_kv.ExistsResponse{Exists: exists}, nil
	}
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
		if val, exists := s.Storage.Get(req.Key); exists {
			s.Storage.Put(req.Key, val, time.Duration(req.Ttl)*time.Second)
			return &pb_kv.ExpireResponse{Success: true, Message: "Expiration Put"}, nil
		}
		return &pb_kv.ExpireResponse{Success: false, Message: "Key not found"}, nil
	}
	value, err := s.Ring.GetWithReplication(req.Key)
	if err != nil {
		return &pb_kv.ExpireResponse{Success: false, Message: "Key not found"}, nil
	}
	err = s.Ring.PutWithReplication(req.Key, value, time.Duration(req.Ttl)*time.Second)
	return &pb_kv.ExpireResponse{Success: err == nil, Message: "Expiration Put"}, nil
}

func (s *OmnidictServer) TTL(ctx context.Context, req *pb_kv.TTLRequest) (*pb_kv.TTLResponse, error) {
	if s.Storage != nil {
		ttl, exists := s.Storage.GetTTL(req.Key)
		if !exists {
			return &pb_kv.TTLResponse{Ttl: -2}, nil
		}
		return &pb_kv.TTLResponse{Ttl: int64(ttl.Seconds())}, nil
	}
	ttl, exists := s.Ring.GetTTL(req.Key)
	if !exists {
		return &pb_kv.TTLResponse{Ttl: -2}, nil
	}
	return &pb_kv.TTLResponse{Ttl: int64(ttl.Seconds())}, nil
}

func (s *OmnidictServer) GetNodeInfo(ctx context.Context, req *pb_kv.NodeInfoRequest) (*pb_kv.NodeInfoResponse, error) {
	nodeNames := s.Ring.GetNodeNames()
	return &pb_kv.NodeInfoResponse{
		NodeId:     "node-1",
		Address:    "localhost:50051",
		Status:     "healthy",
		TotalNodes: int32(len(nodeNames)),
		Nodes:      nodeNames,
	}, nil
}
func NewOmnidictServerWithStorage(s *store.Store) *OmnidictServer {
	return &OmnidictServer{
		Storage: s,
	}
}
