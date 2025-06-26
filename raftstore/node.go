package raftnode

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

// Config holds configuration for a Raft node.
type Config struct {
	NodeID      string // Unique ID of the node (e.g., "node1")
	RaftBind    string // Raft bind address (e.g., ":12000")
	DataDir     string // Directory to store logs/snapshots
	IsBootstrap bool   // True if this is the first node
}

// NewRaftNode sets up and returns a Raft instance.
func NewRaftNode(cfg Config, fsm *FSM, logOutput io.Writer) (*raft.Raft, error) {
	// Create data directories if they don't exist
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create raft dir: %v", err)
	}

	// Set up Raft config
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(cfg.NodeID)
	raftConfig.SnapshotInterval = 20 * time.Second
	raftConfig.SnapshotThreshold = 2

	// Set up log store and stable store
	logStorePath := filepath.Join(cfg.DataDir, "raft-log.bolt")
	stableStore, err := raftboltdb.NewBoltStore(logStorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create bolt store: %v", err)
	}

	// Snapshot store
	snapStore, err := raft.NewFileSnapshotStore(cfg.DataDir, 1, logOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %v", err)
	}

	// Transport layer (binds to TCP port)
	addr, err := raft.NewTCPTransport(cfg.RaftBind, nil, 3, 10*time.Second, logOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %v", err)
	}

	// Initialize Raft system
	r, err := raft.NewRaft(raftConfig, fsm, stableStore, stableStore, snapStore, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize raft: %v", err)
	}

	// Bootstrap cluster if first node
	if cfg.IsBootstrap {
		config := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(cfg.NodeID),
					Address: addr.LocalAddr(),
				},
			},
		}
		r.BootstrapCluster(config)
	}

	return r, nil
}