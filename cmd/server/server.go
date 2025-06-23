package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"omnidict/internal/ring" // consistent hashing logic
	pb "omnidict/proto"      // proto generated Go code

	"google.golang.org/grpc"
)

// remove Test funtion :P

// verifies gRPC server connection and hash ring integration
func (s *server) Test(ctx context.Context, req *pb.TestRequest) (*pb.TestResponse, error) {
	fmt.Printf("🔌 [SERVER] Received test request: step %d, message: %s\n", req.Step, req.Message)
	
	// we've reached the hash ring
	fmt.Printf("🔄 [HASHRING] Test request reached hash ring layer\n")
	fmt.Printf("📍 [HASHRING] Processing step %d: %s\n", req.Step, req.Message)
	
	// hash ring status
	ringStatus := fmt.Sprintf("Hash ring active - Current node: %s", s.selfID)
	
	// test hash ring functionality with a sample key
	testKey := "test_key_integration"
	targetNode := s.ring.GetNode(testKey)
	fmt.Printf("🎯 [HASHRING] Sample key '%s' would route to node: %s\n", testKey, targetNode)
	
	processedMessage := fmt.Sprintf("Hash ring processed: %s", req.Message)
	serverStatus := fmt.Sprintf("Server OK | %s | Target node for test: %s", ringStatus, targetNode)
	
	return &pb.TestResponse{
		Success:      true,
		Message:      processedMessage,
		Step:         req.Step,
		ServerStatus: serverStatus,
	}, nil
}

// server implements the gRPC KVStore service
type server struct {
	pb.UnimplementedKVStoreServer
	data   map[string]string // Local in-memory store
	ring   *ring.HashRing    // Reference to the consistent hashing ring
	selfID string            // Node ID of this server
}

// Put stores the key-value pair
func (s *server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	targetNode := s.ring.GetNode(req.Key)
	if targetNode != s.selfID {
		fmt.Printf("[Redirect] Key '%s' belongs to node '%s'\n", req.Key, targetNode)
		// TODO: forward to appropriate node via gRPC in real implementation
		return &pb.PutResponse{Success: false, Error: "Key belongs to another node"}, nil
	}

	s.data[req.Key] = req.Value
	fmt.Printf("[Stored] Key: %s, Value: %s\n", req.Key, req.Value)
	return &pb.PutResponse{Success: true}, nil
}

// Get retrieves a value by key
func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	targetNode := s.ring.GetNode(req.Key)
	if targetNode != s.selfID {
		fmt.Printf("[Redirect] Key '%s' belongs to node '%s'\n", req.Key, targetNode)
		return &pb.GetResponse{Found: false}, nil
	}

	val, ok := s.data[req.Key]
	if !ok {
		return &pb.GetResponse{Found: false}, nil
	}
	return &pb.GetResponse{Found: true, Value: val}, nil
}

func main() {
	selfID := "node1" // This would be unique per instance
	nodeAddr := ":50051"

	hashRing := ring.NewHashRing(3) // Create a new hash ring with 3 virtual nodes per physical node
	// Add this node to the hash ring
	hashRing.AddNode(selfID)

	s := &server{
		data:   make(map[string]string),
		ring:   hashRing,
		selfID: selfID,
	}

	lis, err := net.Listen("tcp", nodeAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterKVStoreServer(grpcServer, s)

	fmt.Printf("Node %s listening on %s...\n", selfID, nodeAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
