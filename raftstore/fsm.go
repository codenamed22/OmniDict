package raftstore

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	pb_kv "omnidict/proto/kv"
	"omnidict/store"

	"github.com/hashicorp/raft"
)

type Command struct {
	Op    string                `json:"op"`
	Key   string                `json:"key,omitempty"`
	Value string                `json:"value,omitempty"`
	TTL   time.Duration         `json:"ttl,omitempty"`
	TxnID string                `json:"txn_id,omitempty"`
	Ops   []*pb_kv.TxnOperation `json:"ops,omitempty"`
}

type FSM struct {
	mu    sync.Mutex
	store *store.Store
}

func NewFSM(store *store.Store) *FSM {
	return &FSM{
		store: store,
	}
}

func (f *FSM) Apply(l *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	var cmd Command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		log.Println("Failed to unmarshal command:", err)
		return nil
	}

	switch cmd.Op {
	case "put":
		f.store.Put(cmd.Key, cmd.Value, time.Duration(cmd.TTL)*time.Second)
	case "delete":
		f.store.Delete(cmd.Key)
	case "update":
		f.store.Update(cmd.Key, cmd.Value)
	case "expire":
		f.store.Expire(cmd.Key, time.Duration(cmd.TTL)*time.Second)
	case "flush":
		f.store.Flush()
	case "prepare":
		f.store.StageOperations(cmd.TxnID, cmd.Ops)
	case "commit":
		f.store.CommitStagedOperations(cmd.TxnID)
	case "abort":
		f.store.ClearStagedOperations(cmd.TxnID)
	default:
		log.Println("Unknown command op:", cmd.Op)
	}
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	snap := make(map[string]string)
	keys := f.store.GetAllKeys()
	for _, key := range keys {
		if value, exists := f.store.Get(key); exists {
			snap[key] = value
		}
	}
	return &fsmSnapshot{store: snap}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	snap := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&snap); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.store.Flush()
	for key, value := range snap {
		f.store.Put(key, value, 0)
	}
	return nil
}

type fsmSnapshot struct {
	store map[string]string
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	data, err := json.Marshal(s.store)
	if err != nil {
		sink.Cancel()
		return err
	}
	if _, err := sink.Write(data); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *fsmSnapshot) Release() {}
