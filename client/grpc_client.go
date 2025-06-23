package client

import (
	"log"
	"google.golang.org/grpc"
	pb "omnidict/proto"
)

var Conn *grpc.ClientConn
var GrpcClient pb.KVServiceClient

func InitGRPCClient() {
	var err error
	Conn, err = grpc.Dial("localhost:50051", grpc.WithInsecure()) // Change port if needed
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	GrpcClient = pb.NewKVServiceClient(Conn)
}