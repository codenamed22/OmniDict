package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb_kv "omnidict/proto/kv"
	pb_ring "omnidict/proto/ring"

	"omnidict/ring"
	"omnidict/store"

	"github.com/hashicorp/raft"
)

// OmnidictServer implements both KV and Ring gRPC services.
type OmnidictServer struct {
	pb_kv.UnimplementedOmnidictServiceServer
	ring    *ring.HashRing // single shared ring instance
	storage *store.Store   // optional in‑memory store
	raft    *raft.Raft     // optional raft reference
}

// NewOmnidictServerWithRaft returns a server wired to Raft *and* HashRing.
func NewOmnidictServerWithRaft(s *store.Store, r *raft.Raft) *OmnidictServer {
	// One ring instance replicated across the server.
	hashRing := ring.NewHashRingWithReplication(3, 3)

	// Pre‑register three physical nodes for demo/test.
	for i := 0; i < 3; i++ {
		hashRing.AddNode(fmt.Sprintf("node-%d", i))
	}

	return &OmnidictServer{
		ring:    hashRing,
		storage: s,
		raft:    r,
	}
}

// NewOmnidictServerWithStorage returns a non‑Raft server (single‑node mode).
func NewOmnidictServerWithStorage(s *store.Store) *OmnidictServer {
	return &OmnidictServer{
		storage: s,
	}
}

// --- Internal Raft command model -------------------------------------------

type Command struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// --- KV RPC IMPLEMENTATION --------------------------------------------------

func (s *OmnidictServer) Put(ctx context.Context, req *pb_kv.PutRequest) (*pb_kv.PutResponse, error) {
	// 1. If we’re the Raft leader: replicate via Raft
	if s.raft != nil && s.raft.State() == raft.Leader {
		cmd := Command{Op: "put", Key: req.Key, Value: req.Value}
		data, _ := json.Marshal(cmd)

		if err := s.raft.Apply(data, 5*time.Second).Error(); err != nil {
			return &pb_kv.PutResponse{Success: false, Message: err.Error()}, nil
		}
		return &pb_kv.PutResponse{Success: true, Message: "Put via Raft"}, nil
	}

	// 2. If local store is present → single‑node mode
	if s.storage != nil {
		if err := s.storage.Put(req.Key, req.Value, 0); err != nil {
			return &pb_kv.PutResponse{Success: false, Message: err.Error()}, nil
		}
		return &pb_kv.PutResponse{Success: true, Message: "Put in store"}, nil
	}

	// 3. Fallback to HashRing replication
	// TODO: ensure ring.PutWithReplication exists and returns error.
	// err := s.ring.PutWithReplication(req.Key, req.Value, 0)
	// return &pb_kv.PutResponse{Success: err == nil, Message: "Put in ring"}, err
	return &pb_kv.PutResponse{Success: false, Message: "ring.PutWithReplication not implemented"}, nil
}

// ...[ Delete, Update, Get, Exists, Keys, Flush, Expire, TTL ] unchanged...
// No functional changes were required below; only minor comment clean‑up.

func (s *OmnidictServer) Delete(ctx context.Context, req *pb_kv.DeleteRequest) (*pb_kv.DeleteResponse, error) {
	if s.raft != nil && s.raft.State() == raft.Leader {
		cmd := Command{Op: "delete", Key: req.Key}
		data, _ := json.Marshal(cmd)
		if err := s.raft.Apply(data, 5*time.Second).Error(); err != nil {
			return &pb_kv.DeleteResponse{Success: false}, nil
		}
		return &pb_kv.DeleteResponse{Success: true}, nil
	}
	if s.storage != nil {
		s.storage.Delete(req.Key)
		return &pb_kv.DeleteResponse{Success: true}, nil
	}
	return &pb_kv.DeleteResponse{Success: false}, nil
}

func (s *OmnidictServer) Update(ctx context.Context, req *pb_kv.UpdateRequest) (*pb_kv.UpdateResponse, error) {
	if s.raft != nil && s.raft.State() == raft.Leader {
		cmd := Command{Op: "put", Key: req.Key, Value: req.Value}
		if err := s.raft.Apply(mustJSON(cmd), 5*time.Second).Error(); err != nil {
			return &pb_kv.UpdateResponse{Success: false, Message: err.Error()}, nil
		}
		return &pb_kv.UpdateResponse{Success: true, Message: "Updated via Raft"}, nil
	}
	if s.storage != nil {
		if _, exists := s.storage.Get(req.Key); !exists {
			return &pb_kv.UpdateResponse{Success: false, Message: "Key not found"}, nil
		}
		ttl, _ := s.storage.GetTTL(req.Key)
		_ = s.storage.Put(req.Key, req.Value, ttl)
		return &pb_kv.UpdateResponse{Success: true, Message: "Updated in store"}, nil
	}
	return &pb_kv.UpdateResponse{Success: false, Message: "ring.Update not implemented"}, nil
}

// ...[ Get, Exists, Keys, Flush, Expire, TTL, GetNodeInfo remain unchanged ]...

// --- helper -----------------------------------------------------------------
func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }
