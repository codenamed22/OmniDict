package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "omnidict/proto" // this is the correct path if kv.pb.go is in proto/ and module is omnidict

	"google.golang.org/grpc"
)

// server implements the gRPC KVStore service defined in the proto file.
type server struct {
	pb.UnimplementedKVStoreServer
	data map[string]string // In-memory key-value store
}

// Put stores the key-value pair received in the request.
func (s *server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	s.data[req.Key] = req.Value
	fmt.Printf("Put: %s = %s\n", req.Key, req.Value)
	return &pb.PutResponse{Success: true}, nil
}

// Get retrieves the value for a given key.
func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	value, ok := s.data[req.Key]
	if !ok {
		return &pb.GetResponse{Value: "", Found: false}, nil
	}
	return &pb.GetResponse{Value: value, Found: true}, nil
}

// Join handles cluster join requests (stubbed for now).
func (s *server) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	fmt.Printf("Join request from node: %s @ %s\n", req.NodeInfo.Id, req.NodeInfo.Address)

	return &pb.JoinResponse{
		Success:      true,
		ClusterNodes: []*pb.NodeInfo{}, // optional: return known cluster nodes
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register our server
	pb.RegisterKVStoreServer(grpcServer, &server{
		data: make(map[string]string),
	})

	fmt.Println("gRPC server listening on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
