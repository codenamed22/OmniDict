package client

import (
	"context"
	"log"

	pb_kv "omnidict/proto/kv"

	"google.golang.org/grpc"
)

var (
	globalClient *Client
)

type Transaction struct {
	ID     string
	Client pb_kv.OmnidictServiceClient
	Ops    []*pb_kv.TxnOperation
}

type Client struct {
	conn     *grpc.ClientConn
	kvClient pb_kv.OmnidictServiceClient
}

func InitGRPCClient(addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	globalClient = &Client{
		conn:     conn,
		kvClient: pb_kv.NewOmnidictServiceClient(conn),
	}
}

func Close() {
	if globalClient != nil && globalClient.conn != nil {
		globalClient.conn.Close()
	}
}

func Put(key string, value []byte) (*pb_kv.PutResponse, error) {
	return globalClient.kvClient.Put(context.Background(), &pb_kv.PutRequest{
		Key:   key,
		Value: string(value),
	})
}

func Get(key string) (*pb_kv.GetResponse, error) {
	return globalClient.kvClient.Get(context.Background(), &pb_kv.GetRequest{Key: key})
}

func Delete(key string) (*pb_kv.DeleteResponse, error) {
	return globalClient.kvClient.Delete(context.Background(), &pb_kv.DeleteRequest{Key: key})
}

func Exists(key string) (*pb_kv.ExistsResponse, error) {
	return globalClient.kvClient.Exists(context.Background(), &pb_kv.ExistsRequest{Key: key})
}

func Expire(key string, ttl int64) (*pb_kv.ExpireResponse, error) {
	return globalClient.kvClient.Expire(context.Background(), &pb_kv.ExpireRequest{
		Key: key,
		Ttl: ttl,
	})
}

func Flush() (*pb_kv.FlushResponse, error) {
	return globalClient.kvClient.Flush(context.Background(), &pb_kv.FlushRequest{})
}

func Keys(pattern string) (*pb_kv.KeysResponse, error) {
	return globalClient.kvClient.Keys(context.Background(), &pb_kv.KeysRequest{Pattern: pattern})
}

func TTL(key string) (*pb_kv.TTLResponse, error) {
	return globalClient.kvClient.TTL(context.Background(), &pb_kv.TTLRequest{Key: key})
}

func Update(key string, value []byte) (*pb_kv.UpdateResponse, error) {
	return globalClient.kvClient.Update(context.Background(), &pb_kv.UpdateRequest{
		Key:   key,
		Value: string(value),
	})
}

func BeginTransaction() (*Transaction, error) {
	resp, err := globalClient.kvClient.BeginTransaction(context.Background(), &pb_kv.BeginTxnRequest{})
	if err != nil {
		return nil, err
	}
	return &Transaction{
		ID:     resp.TxnId,
		Client: globalClient.kvClient,
	}, nil
}

func (t *Transaction) Put(key, value string) {
	t.Ops = append(t.Ops, &pb_kv.TxnOperation{
		Key:   key,
		Value: value,
		Op:    pb_kv.TxnOperation_SET,
	})
}

func (t *Transaction) Delete(key string) {
	t.Ops = append(t.Ops, &pb_kv.TxnOperation{
		Key: key,
		Op:  pb_kv.TxnOperation_DELETE,
	})
}

func (t *Transaction) Commit() (bool, error) {
	// Prepare phase
	prepareResp, err := t.Client.Prepare(context.Background(), &pb_kv.PrepareRequest{
		TxnId:      t.ID,
		Operations: t.Ops,
	})
	if err != nil || !prepareResp.Success {
		t.Client.Abort(context.Background(), &pb_kv.AbortRequest{TxnId: t.ID})
		return false, err
	}

	// Commit phase
	commitResp, err := t.Client.Commit(context.Background(), &pb_kv.CommitRequest{TxnId: t.ID})
	if err != nil {
		return false, err
	}
	return commitResp.Success, nil
}

func (t *Transaction) Abort() (bool, error) {
	abortResp, err := t.Client.Abort(context.Background(), &pb_kv.AbortRequest{TxnId: t.ID})
	if err != nil {
		return false, err
	}
	return abortResp.Success, nil
}
