package client

import (
	"context"
	"fmt"
	"log"
	"google.golang.org/grpc"
	pb "omnidict/proto"
)

var Conn *grpc.ClientConn
var GrpcClient pb.KVStoreClient

func InitGRPCClient() {
	var err error
	Conn, err = grpc.Dial("localhost:50051", grpc.WithInsecure()) // Change port if needed
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	GrpcClient = pb.NewKVStoreClient(Conn)
}

// TestConnection - Wrapper function to show client layer processing
func TestConnection(message string, step int32) (*pb.TestResponse, error) {
	fmt.Printf("📡 [CLIENT] Preparing test request: step %d, message: %s\n", step, message)
	fmt.Printf("🔗 [CLIENT] Sending gRPC request to server...\n")
	
	// Make the actual gRPC call
	resp, err := GrpcClient.Test(context.Background(), &pb.TestRequest{
		Message: message,
		Step:    step,
	})
	
	if err != nil {
		fmt.Printf("❌ [CLIENT] gRPC call failed: %v\n", err)
		return nil, err
	}
	
	fmt.Printf("✅ [CLIENT] Received response from server\n")
	fmt.Printf("📨 [CLIENT] Server response: %s\n", resp.Message)
	
	return resp, nil
}