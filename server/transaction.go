package server

import (
	"context"
	"sync"
	"time"

	pb_kv "omnidict/proto/kv"
	"omnidict/store"

	"github.com/hashicorp/raft"
)

type TransactionCoordinator struct {
	mu           sync.Mutex
	transactions map[string]*Transaction
	shardRafts   map[string]*raft.Raft
	shardStores  map[string]*store.Store
}

type Transaction struct {
	ID  string
	Ops []*pb_kv.TxnOperation
}

func NewTransactionCoordinator(shardRafts map[string]*raft.Raft, shardStores map[string]*store.Store) *TransactionCoordinator {
	return &TransactionCoordinator{
		transactions: make(map[string]*Transaction),
		shardRafts:   shardRafts,
		shardStores:  shardStores,
	}
}

func (tc *TransactionCoordinator) BeginTransaction(ctx context.Context, req *pb_kv.BeginTxnRequest) (*pb_kv.BeginTxnResponse, error) {
	txnID := generateTxnID()
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.transactions[txnID] = &Transaction{
		ID: txnID,
	}
	return &pb_kv.BeginTxnResponse{TxnId: txnID}, nil
}

func (tc *TransactionCoordinator) Prepare(ctx context.Context, req *pb_kv.PrepareRequest) (*pb_kv.PrepareResponse, error) {
	return &pb_kv.PrepareResponse{Success: true}, nil
}

func (tc *TransactionCoordinator) Commit(ctx context.Context, req *pb_kv.CommitRequest) (*pb_kv.CommitResponse, error) {
	return &pb_kv.CommitResponse{Success: true}, nil
}

func (tc *TransactionCoordinator) Abort(ctx context.Context, req *pb_kv.AbortRequest) (*pb_kv.AbortResponse, error) {
	return &pb_kv.AbortResponse{Success: true}, nil
}

func generateTxnID() string {
	return "txn-" + time.Now().Format("20060102150405")
}
