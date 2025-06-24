package client

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "omnidict/proto"
)

var Conn *grpc.ClientConn

func InitGRPCClient(endpoint string) {
	var err error
	Conn, err = grpc.Dial(endpoint, 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
}

func Put(key string, value []byte) (*pb.PutResponse, error) {
	client := pb.NewRingServiceClient(Conn)
	return client.Put(context.Background(), &pb.PutRequest{
		Key:   key,
		Value: value,
	})
}

func Get(key string) (*pb.GetResponse, error) {
	client := pb.NewRingServiceClient(Conn)
	return client.Get(context.Background(), &pb.GetRequest{Key: key})
}

func Delete(key string) (*pb.DeleteResponse, error) {
	client := pb.NewRingServiceClient(Conn)
	return client.Delete(context.Background(), &pb.DeleteRequest{Key: key})
}
