package client

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "omnidict/proto"
)

var Conn *grpc.ClientConn
var Client pb.OmnidictServiceClient

func InitGRPCClient() {
	serverAddr := "localhost:50051"
	conn, err := grpc.Dial(serverAddr, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	Conn = conn
	Client = pb.NewOmnidictServiceClient(Conn)
}
