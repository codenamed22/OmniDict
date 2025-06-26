package server

import (
	"context"
	"log"
	"net"
	"time"

	pb_kv "omnidict/proto/kv"
	// pb_ring "omnidict/proto/ring" will be used later
	"omnidict/ring"
	// "omnidict/store" later

	"google.golang.org/grpc"
)

type OmnidictServer struct {
	pb_kv.UnimplementedOmnidictServiceServer
	Ring *ring.HashRing
}

func NewOmnidictServer() *OmnidictServer {
	return &OmnidictServer{
		Ring: ring.NewHashRingWithReplication(3, 3),
	}
}

func (s *OmnidictServer) Put(ctx context.Context, req *pb_kv.PutRequest) (*pb_kv.PutResponse, error) {
	// Using 0 TTL for put - no expiration
	err := s.Ring.SetWithReplication(req.Key, req.Value, 0)
	if err != nil {
		return &pb_kv.PutResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb_kv.PutResponse{Success: true, Message: "OK"}, nil
}

func (s *OmnidictServer) Get(ctx context.Context, req *pb_kv.GetRequest) (*pb_kv.GetResponse, error) {
	value, err := s.Ring.GetWithReplication(req.Key)
	if err != nil {
		return &pb_kv.GetResponse{Found: false, Value: ""}, nil
	}
	return &pb_kv.GetResponse{Found: true, Value: value}, nil
}

func (s *OmnidictServer) Delete(ctx context.Context, req *pb_kv.DeleteRequest) (*pb_kv.DeleteResponse, error) {
	err := s.Ring.Delete(req.Key)
	return &pb_kv.DeleteResponse{Success: err == nil}, err
}

func (s *OmnidictServer) Exists(ctx context.Context, req *pb_kv.ExistsRequest) (*pb_kv.ExistsResponse, error) {
	exists := s.Ring.Exists(req.Key)
	return &pb_kv.ExistsResponse{Exists: exists}, nil
}

func (s *OmnidictServer) Keys(ctx context.Context, req *pb_kv.KeysRequest) (*pb_kv.KeysResponse, error) {
	var keys []string
	if req.Pattern == "" {
		keys = s.Ring.GetAllKeys()
	} else {
		keys = s.Ring.GetKeysWithPrefix(req.Pattern)
	}
	return &pb_kv.KeysResponse{Keys: keys}, nil
}

func (s *OmnidictServer) Flush(ctx context.Context, req *pb_kv.FlushRequest) (*pb_kv.FlushResponse, error) {
	// Implement using your hash ring's capabilities
	for _, node := range s.Ring.GetNodeNames() {
		store := s.Ring.GetNodeStore(node)
		for key := range store {
			s.Ring.Delete(key)
		}
	}
	return &pb_kv.FlushResponse{Success: true}, nil
}

func (s *OmnidictServer) Expire(ctx context.Context, req *pb_kv.ExpireRequest) (*pb_kv.ExpireResponse, error) {
	ttl := req.Ttl
	
	// Get current value to preserve it
	value, err := s.Ring.GetWithReplication(req.Key)
	if err != nil {
		return &pb_kv.ExpireResponse{Success: false, Message: "Key not found"}, nil
	}
	
	// Update with TTL
	err = s.Ring.SetWithReplication(req.Key, value, time.Duration(ttl)*time.Second)
	if err != nil {
		return &pb_kv.ExpireResponse{Success: false, Message: err.Error()}, nil
	}
	
	return &pb_kv.ExpireResponse{Success: true, Message: "Expiration set"}, nil
}

func (s *OmnidictServer) TTL(ctx context.Context, req *pb_kv.TTLRequest) (*pb_kv.TTLResponse, error) {
	ttl, exists := s.Ring.GetTTL(req.Key)
	if !exists {
		return &pb_kv.TTLResponse{Ttl: -2}, nil // Key doesn't exist
	}
	return &pb_kv.TTLResponse{Ttl: int64(ttl.Seconds())}, nil
}

func (s *OmnidictServer) Update(ctx context.Context, req *pb_kv.UpdateRequest) (*pb_kv.UpdateResponse, error) {
	if !s.Ring.Exists(req.Key) {
		return &pb_kv.UpdateResponse{Success: false, Message: "Key not found"}, nil
	}
	
	// Preserve existing TTL
	ttl, _ := s.Ring.GetTTL(req.Key)
	err := s.Ring.SetWithReplication(req.Key, req.Value, ttl)
	if err != nil {
		return &pb_kv.UpdateResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb_kv.UpdateResponse{Success: true, Message: "Updated successfully"}, nil
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