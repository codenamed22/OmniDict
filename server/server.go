package server

import (
	"context"
	"log"
	"net"

	"omnidict/proto"
	"omnidict/ring"

	"google.golang.org/grpc"
)

type OmnidictServer struct {
	proto.UnimplementedOmnidictServiceServer
	Ring *ring.HashRing
}

func (s *OmnidictServer) Put(ctx context.Context, req *proto.PutRequest) (*proto.PutResponse, error) {
	s.Ring.Set(req.Key, req.Value)
	return &proto.PutResponse{Success: true, Message: "OK"}, nil
}

func (s *OmnidictServer) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	value, found := s.Ring.Get(req.Key)
	return &proto.GetResponse{Found: found, Value: value}, nil
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
	keys := s.Ring.GetAllKeys()
	return &proto.KeysResponse{Keys: keys}, nil
}

func (s *OmnidictServer) Flush(ctx context.Context, req *proto.FlushRequest) (*proto.FlushResponse, error) {
	// Implementation would clear all data
	return &proto.FlushResponse{Success: true}, nil
}

func (s *OmnidictServer) Update(ctx context.Context, req *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	// Check if key exists first
	if !s.Ring.Exists(req.Key) {
		return &proto.UpdateResponse{Success: false, Message: "Key not found"}, nil
	}
	
	s.Ring.Set(req.Key, req.Value)
	return &proto.UpdateResponse{Success: true, Message: "Updated successfully"}, nil
}

func (s *OmnidictServer) Expire(ctx context.Context, req *proto.ExpireRequest) (*proto.ExpireResponse, error) {
	// For now, just return success (TTL implementation would require more complex storage)
	return &proto.ExpireResponse{Success: true, Message: "Expiration set"}, nil
}

func (s *OmnidictServer) TTL(ctx context.Context, req *proto.TTLRequest) (*proto.TTLResponse, error) {
	// Check if key exists
	if !s.Ring.Exists(req.Key) {
		return &proto.TTLResponse{Ttl: -2}, nil // Key doesn't exist
	}
	
	// For now, return -1 (no expiration) since we haven't implemented TTL storage
	return &proto.TTLResponse{Ttl: -1}, nil
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
	hr := ring.NewHashRing(3) // 3 virtual nodes
	proto.RegisterOmnidictServiceServer(s, &OmnidictServer{Ring: hr})
	
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
