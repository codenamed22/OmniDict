
package raftnode

import (
	"encoding/json"
	"io"
	"log"
	"sync"

	raft "github.com/hashicorp/raft"
)

// Command represents a Raft command with correct struct tags.
type Command struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// FSM represents the finite state machine for Raft.
type FSM struct {
	mu    sync.Mutex
	store map[string]string
}

// NewFSM creates a new FSM.
func NewFSM() *FSM {
	return &FSM{
		store: make(map[string]string),
	}
}

// Apply applies a Raft log entry to the FSM.
func (f *FSM) Apply(l *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		log.Println("Failed to unmarshal command:", err)
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Op {
	case "put":
		f.store[cmd.Key] = cmd.Value
	case "delete":
		delete(f.store, cmd.Key)
	default:
		log.Println("Unknown command op:", cmd.Op)
	}
	return nil
}

// Snapshot returns a snapshot of the FSM.
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	snap := make(map[string]string)
	for k, v := range f.store {
		snap[k] = v
	}
	return &fsmSnapshot{store: snap}, nil
}

// Restore sets the FSM state from a snapshot.
func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	snap := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&snap); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.store = snap
	return nil
}

// fsmSnapshot implements raft.FSMSnapshot
type fsmSnapshot struct {
	store map[string]string
}

// Persist writes the snapshot to a sink.
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

// Release is a no-op for our snapshot.
func (s *fsmSnapshot) Release() {}

