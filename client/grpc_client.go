package client

import (
	"context"

	pb_kv "omnidict/proto/kv"

	"google.golang.org/grpc"
)

var (
	// Global connection and client - change to struct / dependency injection (idk what that is lol T-T)
	Conn     *grpc.ClientConn
	KvClient pb_kv.OmnidictServiceClient
)

func InitGRPCClient(addr string) error {
	var err error
	Conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	KvClient = pb_kv.NewOmnidictServiceClient(Conn)
	return nil
}

func Close() {
	if Conn != nil {
		Conn.Close()
	}
}

// Key-Value Operations
func Put(key string, value []byte) (*pb_kv.PutResponse, error) {
	return KvClient.Put(context.Background(), &pb_kv.PutRequest{
		Key:   key,
		Value: string(value),
	})
}

func Get(key string) (*pb_kv.GetResponse, error) {
	return KvClient.Get(context.Background(), &pb_kv.GetRequest{Key: key})
}

func Delete(key string) (*pb_kv.DeleteResponse, error) {
	return KvClient.Delete(context.Background(), &pb_kv.DeleteRequest{Key: key})
}

func Exists(key string) (*pb_kv.ExistsResponse, error) {
	return KvClient.Exists(context.Background(), &pb_kv.ExistsRequest{Key: key})
}

func Expire(key string, ttl int64) (*pb_kv.ExpireResponse, error) {
	return KvClient.Expire(context.Background(), &pb_kv.ExpireRequest{
		Key: key,
		Ttl: ttl,
	})
}

func Flush() (*pb_kv.FlushResponse, error) {
	return KvClient.Flush(context.Background(), &pb_kv.FlushRequest{})
}

func Keys(pattern string) (*pb_kv.KeysResponse, error) {
	return KvClient.Keys(context.Background(), &pb_kv.KeysRequest{Pattern: pattern})
}

func TTL(key string) (*pb_kv.TTLResponse, error) {
	return KvClient.TTL(context.Background(), &pb_kv.TTLRequest{Key: key})
}

func Update(key string, value []byte) (*pb_kv.UpdateResponse, error) {
	return KvClient.Update(context.Background(), &pb_kv.UpdateRequest{
		Key:   key,
		Value: string(value),
	})
}

// Transaction Handling - move this later to a separate file
type Transaction struct {
	ID  string
	Ops []*pb_kv.TxnOperation
}

func BeginTransaction() (*Transaction, error) {
	resp, err := KvClient.BeginTransaction(context.Background(), &pb_kv.BeginTxnRequest{})
	if err != nil {
		return nil, err
	}
	return &Transaction{ID: resp.TxnId}, nil
}

func (t *Transaction) AddOperation(opType pb_kv.TxnOperation_TxnOp, key, value string) {
	t.Ops = append(t.Ops, &pb_kv.TxnOperation{
		Key:   key,
		Value: value,
		Op:    opType,
	})
}

func (t *Transaction) Put(key, value string) {
	t.AddOperation(pb_kv.TxnOperation_SET, key, value)
}

func (t *Transaction) Delete(key string) {
	t.AddOperation(pb_kv.TxnOperation_DELETE, key, "")
}

func (t *Transaction) Commit() error {
	// Prepare phase
	_, err := KvClient.Prepare(context.Background(), &pb_kv.PrepareRequest{
		TxnId:      t.ID,
		Operations: t.Ops,
	})
	if err != nil {
		t.Abort()
		return err
	}

	// Commit phase
	_, err = KvClient.Commit(context.Background(), &pb_kv.CommitRequest{TxnId: t.ID})
	return err
}

func (t *Transaction) Abort() error {
	_, err := KvClient.Abort(context.Background(), &pb_kv.AbortRequest{TxnId: t.ID})
	return err
}
