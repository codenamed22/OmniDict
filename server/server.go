package server

import (
	"context"
	"log"
	"net"
	"time"

	"omnidict/proto"
	"omnidict/ring"

	"google.golang.org/grpc"
)

type OmnidictServer struct {
	proto.UnimplementedOmnidictServiceServer
	Ring *ring.HashRing
}

func NewOmnidictServer() *OmnidictServer {
	return &OmnidictServer{
		Ring: ring.NewHashRingWithReplication(3, 3),
	}
}

func (s *OmnidictServer) Put(ctx context.Context, req *proto.PutRequest) (*proto.PutResponse, error) {
	// Using 0 TTL for put (no expiration)
	err := s.Ring.SetWithReplication(req.Key, req.Value, 0)
	if err != nil {
		return &proto.PutResponse{Success: false, Message: err.Error()}, nil
	}
	return &proto.PutResponse{Success: true, Message: "OK"}, nil
}

func (s *OmnidictServer) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	value, err := s.Ring.GetWithReplication(req.Key)
	if err != nil {
		return &proto.GetResponse{Found: false, Value: ""}, nil
	}
	return &proto.GetResponse{Found: true, Value: value}, nil
}

func (s *OmnidictServer) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	err := s.Ring.Delete(req.Key)
	return &proto.DeleteResponse{Success: err == nil}, err
}

func (s *OmnidictServer) Exists(ctx context.Context, req *proto.ExistsRequest) (*proto.ExistsResponse, error) {
	exists := s.Ring.Exists(req.Key)
	return &proto.ExistsResponse{Exists: exists}, nil
}

func (s *OmnidictServer) Keys(ctx context.Context, req *proto.KeysRequest) (*proto.KeysResponse, error) {
	var keys []string
	if req.Pattern == "" {
		keys = s.Ring.GetAllKeys()
	} else {
		keys = s.Ring.GetKeysWithPrefix(req.Pattern)
	}
	return &proto.KeysResponse{Keys: keys}, nil
}

func (s *OmnidictServer) Flush(ctx context.Context, req *proto.FlushRequest) (*proto.FlushResponse, error) {
	// Implement using your hash ring's capabilities
	for _, node := range s.Ring.GetNodeNames() {
		store := s.Ring.GetNodeStore(node)
		for key := range store {
			s.Ring.Delete(key)
		}
	}
	return &proto.FlushResponse{Success: true}, nil
}

func (s *OmnidictServer) Expire(ctx context.Context, req *proto.ExpireRequest) (*proto.ExpireResponse, error) {
	ttl := req.Ttl
	
	// Get current value to preserve it
	value, err := s.Ring.GetWithReplication(req.Key)
	if err != nil {
		return &proto.ExpireResponse{Success: false, Message: "Key not found"}, nil
	}
	
	// Update with TTL
	err = s.Ring.SetWithReplication(req.Key, value, time.Duration(ttl)*time.Second)
	if err != nil {
		return &proto.ExpireResponse{Success: false, Message: err.Error()}, nil
	}
	
	return &proto.ExpireResponse{Success: true, Message: "Expiration set"}, nil
}

func (s *OmnidictServer) TTL(ctx context.Context, req *proto.TTLRequest) (*proto.TTLResponse, error) {
	ttl, exists := s.Ring.GetTTL(req.Key)
	if !exists {
		return &proto.TTLResponse{Ttl: -2}, nil // Key doesn't exist
	}
	return &proto.TTLResponse{Ttl: int64(ttl.Seconds())}, nil
}

func (s *OmnidictServer) Update(ctx context.Context, req *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	if !s.Ring.Exists(req.Key) {
		return &proto.UpdateResponse{Success: false, Message: "Key not found"}, nil
	}
	
	// Preserve existing TTL
	ttl, _ := s.Ring.GetTTL(req.Key)
	err := s.Ring.SetWithReplication(req.Key, req.Value, ttl)
	if err != nil {
		return &proto.UpdateResponse{Success: false, Message: err.Error()}, nil
	}
	return &proto.UpdateResponse{Success: true, Message: "Updated successfully"}, nil
}

func (s *OmnidictServer) GetNodeInfo(ctx context.Context, req *proto.NodeInfoRequest) (*proto.NodeInfoResponse, error) {
	nodeNames := s.Ring.GetNodeNames()
	return &proto.NodeInfoResponse{
		NodeId:     "node-1",
		Address:    "localhost:50051",
		Status:     "healthy",
		TotalNodes: int32(len(nodeNames)),
		Nodes:      nodeNames,
	}, nil
}

func StartServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	server := NewOmnidictServer()
	proto.RegisterOmnidictServiceServer(s, server)

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
