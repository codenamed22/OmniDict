
package raftnode

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"
	"strings"

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
	mu    sync.RWMutex
	store map[string]string
	expiry map[string]time.Time
}

// NewFSM creates a new FSM.
func NewFSM() *FSM {
	return &FSM{
		store: make(map[string]string),
		expiry: make(map[string]time.Time),
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
func (f *FSM) Exists(key string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, ok := f.store[key]
	return ok, nil
}
func (f *FSM) Expire(key string, ttl int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Sample logic: Add expiry logic in your store
	// This could be a map[string]time.Time or map[string]int64 with timestamps
	expiryTime := time.Now().Add(time.Duration(ttl) * time.Second)
	f.expiry[key] = expiryTime

	return nil
}
func (f *FSM) TTL(key string) (int64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Does the key exist?
	_, exists := f.store[key]
	if !exists {
		return -2, nil // Key doesn't exist
	}

	expiry, ok := f.expiry[key]
	if !ok {
		return -1, nil // Key has no expiration
	}

	remaining := int64(time.Until(expiry).Seconds())
	if remaining <= 0 {
		return -2, nil // Considered expired
	}

	return remaining, nil
}
func (f *FSM) Flush() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Clear the store and expiry map
	f.store = make(map[string]string)
	f.expiry = make(map[string]time.Time)

	return nil
}
func (f *FSM) Get(key string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Optional: Check if key has expired
	if expiryTime, ok := f.expiry[key]; ok {
		if time.Now().After(expiryTime) {
			// Treat as not found
			delete(f.store, key)   // Optional cleanup
			delete(f.expiry, key) // Optional cleanup
			return "", false
		}
	}

	value, ok := f.store[key]
	return value, ok
}
func (f *FSM) Keys(pattern string) ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	now := time.Now()
	var result []string

	for key := range f.store {
		// ❌ Skip expired keys
		if expiryTime, ok := f.expiry[key]; ok {
			if now.After(expiryTime) {
				continue
			}
		}

		// ✅ Match pattern (simple substring match)
		if pattern == "" || strings.Contains(key, pattern) {
			result = append(result, key)
		}
	}

	return result, nil
}
switch cmd.Op {
case "put":
	f.store[cmd.Key] = string(cmd.Value)
	delete(f.expiry, cmd.Key)

case "update":
	if _, exists := f.store[cmd.Key]; exists {
		f.store[cmd.Key] = string(cmd.Value)
		delete(f.expiry, cmd.Key)
	}
}