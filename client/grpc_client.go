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
	Conn, err = grpc.Dial(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
}

func Put(key string, value []byte) (*pb_kv.PutResponse, error) {
	client := pb_kv.NewOmnidictServiceClient(Conn) // Fixed service name
	return client.Put(context.Background(), &pb_kv.PutRequest{
		Key:   key,
		Value: string(value),
	})
}

func Get(key string) (*pb_kv.GetResponse, error) {
	client := pb_kv.NewOmnidictServiceClient(Conn)
	return client.Get(context.Background(), &pb_kv.GetRequest{Key: key})
}

func Delete(key string) (*pb_kv.DeleteResponse, error) {
	client := pb_kv.NewOmnidictServiceClient(Conn)
	return client.Delete(context.Background(), &pb_kv.DeleteRequest{Key: key})
}

func Exists(key string) (*pb_kv.ExistsResponse, error) {
	client := pb_kv.NewOmnidictServiceClient(Conn)
	return client.Exists(context.Background(), &pb_kv.ExistsRequest{Key: key})
}

func Expire(key string, ttl int64) (*pb_kv.ExpireResponse, error) {
	client := pb_kv.NewOmnidictServiceClient(Conn)
	return client.Expire(context.Background(), &pb_kv.ExpireRequest{
		Key: key,
		Ttl: ttl,
	})
}

func Flush() (*pb_kv.FlushResponse, error) {
	client := pb_kv.NewOmnidictServiceClient(Conn)
	return client.Flush(context.Background(), &pb_kv.FlushRequest{})
}

func Keys(pattern string) (*pb_kv.KeysResponse, error) {
	client := pb_kv.NewOmnidictServiceClient(Conn)
	return client.Keys(context.Background(), &pb_kv.KeysRequest{Pattern: pattern})
}

func TTL(key string) (*pb_kv.TTLResponse, error) {
	client := pb_kv.NewOmnidictServiceClient(Conn)
	return client.TTL(context.Background(), &pb_kv.TTLRequest{Key: key})
}

func Update(key string, value []byte) (*pb_kv.UpdateResponse, error) {
	client := pb_kv.NewOmnidictServiceClient(Conn)
	return client.Update(context.Background(), &pb_kv.UpdateRequest{
		Key:   key,
		Value: string(value),
	})
}