package client

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb_kv "omnidict/proto/kv"
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

func Put(key string, value []byte) (*pb_kv.PutResponse, error) {
	client := pb_kv.NewRingServiceClient(Conn)
	return client.Put(context.Background(), &pb_kv.PutRequest{
		Key:   key,
		Value: value,
	})
}

func Get(key string) (*pb_kv.GetResponse, error) {
	client := pb_kv.NewRingServiceClient(Conn)
	return client.Get(context.Background(), &pb_kv.GetRequest{Key: key})
}

func Delete(key string) (*pb_kv.DeleteResponse, error) {
	client := pb_kv.NewRingServiceClient(Conn)
	return client.Delete(context.Background(), &pb_kv.DeleteRequest{Key: key})
}
