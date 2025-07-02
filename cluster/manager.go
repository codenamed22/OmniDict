package cluster

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"omnidict/raftstore"
	"omnidict/store"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

const numShards = 3

type ShardManager struct {
	nodeID   string
	dataDir  string
	raftPort string
}

func NewShardManager(nodeID, dataDir, raftPort string) *ShardManager {
	return &ShardManager{
		nodeID:   nodeID,
		dataDir:  dataDir,
		raftPort: raftPort,
	}
}

func (sm *ShardManager) InitializeShards() (map[string]*raft.Raft, map[string]*store.Store, error) {
	shardRafts := make(map[string]*raft.Raft)
	shardStores := make(map[string]*store.Store)
	basePort, _ := strconv.Atoi(sm.raftPort)

	var lastErr error
	successCount := 0

	for i := 0; i < numShards; i++ {
		shardID := "shard" + strconv.Itoa(i)
		shardPort := strconv.Itoa(basePort + i)
		shardDataDir := filepath.Join(sm.dataDir, shardID)

		r, kvStore := sm.createRaftForShard(shardID, shardDataDir, shardPort)
		if r != nil && kvStore != nil {
			shardRafts[shardID] = r
			shardStores[shardID] = kvStore
			successCount++
			log.Printf("Initialized shard %s with Raft port %s", shardID, shardPort)
		} else {
			lastErr = fmt.Errorf("failed to initialize shard %s", shardID)
			log.Printf("Failed to initialize shard %s", shardID)
		}
	}

	// Return error if no shards were successfully initialized
	if successCount == 0 {
		return nil, nil, fmt.Errorf("failed to initialize any shards: %v", lastErr)
	}

	// Return warning if some shards failed but some succeeded
	if successCount < numShards {
		return shardRafts, shardStores, fmt.Errorf("only %d/%d shards initialized successfully", successCount, numShards)
	}

	return shardRafts, shardStores, nil
}

func (sm *ShardManager) createRaftForShard(shardID, dataDir, raftPort string) (*raft.Raft, *store.Store) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(sm.nodeID + "-" + shardID)

	// Create data directory
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		log.Printf("[%s] failed to create data directory: %v", shardID, err)
		return nil, nil
	}

	// Create BoltDB stores
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.bolt"))
	if err != nil {
		log.Printf("[%s] failed to create log store: %v", shardID, err)
		return nil, nil
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.bolt"))
	if err != nil {
		log.Printf("[%s] failed to create stable store: %v", shardID, err)
		logStore.Close() // Clean up
		return nil, nil
	}

	snapshotStore, err := raft.NewFileSnapshotStore(dataDir, 1, os.Stderr)
	if err != nil {
		log.Printf("[%s] failed to create snapshot store: %v", shardID, err)
		logStore.Close()
		stableStore.Close()
		return nil, nil
	}

	// Create TCP transport
	addr := net.JoinHostPort("127.0.0.1", raftPort)
	transport, err := raft.NewTCPTransport(addr, nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		log.Printf("[%s] failed to create transport: %v", shardID, err)
		logStore.Close()
		stableStore.Close()
		return nil, nil
	}

	// Create store and FSM
	kvStore := store.NewStore()
	fsm := raftstore.NewFSM(kvStore)

	// Check if cluster already exists
	hasState, err := raft.HasExistingState(logStore, stableStore, snapshotStore)
	if err != nil {
		log.Printf("[%s] failed to check existing state: %v", shardID, err)
		logStore.Close()
		stableStore.Close()
		transport.Close()
		return nil, nil
	}

	// Create Raft instance
	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		log.Printf("[%s] failed to start Raft node: %v", shardID, err)
		logStore.Close()
		stableStore.Close()
		transport.Close()
		return nil, nil
	}

	// Bootstrap single-node cluster only if no existing state
	if !hasState {
		raftConfig := raft.Configuration{
			Servers: []raft.Server{{
				ID:      config.LocalID,
				Address: transport.LocalAddr(),
			}},
		}

		bootstrapFuture := r.BootstrapCluster(raftConfig)
		if err := bootstrapFuture.Error(); err != nil {
			log.Printf("[%s] failed to bootstrap cluster: %v", shardID, err)
			r.Shutdown()
			return nil, nil
		}
		log.Printf("[%s] Bootstrapped Raft cluster at %s", shardID, addr)
	} else {
		log.Printf("[%s] Joined existing Raft cluster at %s", shardID, addr)
	}

	return r, kvStore
}
