package cluster

import (
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

func (sm *ShardManager) InitializeShards() (map[string]*raft.Raft, map[string]*store.Store) {
	shardRafts := make(map[string]*raft.Raft)
	shardStores := make(map[string]*store.Store)
	basePort, _ := strconv.Atoi(sm.raftPort)

	for i := 0; i < numShards; i++ {
		shardID := "shard" + strconv.Itoa(i)
		shardPort := strconv.Itoa(basePort + i)
		shardDataDir := filepath.Join(sm.dataDir, shardID)

		r, store := sm.createRaftForShard(shardID, shardDataDir, shardPort)
		shardRafts[shardID] = r
		shardStores[shardID] = store
		log.Printf("Initialized shard %s with Raft port %s", shardID, shardPort)
	}

	return shardRafts, shardStores
}

func (sm *ShardManager) createRaftForShard(shardID, dataDir, raftPort string) (*raft.Raft, *store.Store) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(sm.nodeID + "-" + shardID)
	os.MkdirAll(dataDir, 0700)

	// Create BoltDB stores
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.bolt"))
	if err != nil {
		log.Fatalf("[%s] failed to create log store: %v", shardID, err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.bolt"))
	if err != nil {
		log.Fatalf("[%s] failed to create stable store: %v", shardID, err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(dataDir, 1, os.Stderr)
	if err != nil {
		log.Fatalf("[%s] failed to create snapshot store: %v", shardID, err)
	}

	// Create TCP transport
	addr := net.JoinHostPort("127.0.0.1", raftPort)
	transport, err := raft.NewTCPTransport(addr, nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		log.Fatalf("[%s] failed to create transport: %v", shardID, err)
	}

	// Create store and FSM
	kvStore := store.NewStore()
	fsm := raftstore.NewFSM(kvStore)

	// Create Raft instance
	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		log.Fatalf("[%s] failed to start Raft node: %v", shardID, err)
	}

	// Bootstrap single-node cluster
	raftConfig := raft.Configuration{
		Servers: []raft.Server{{
			ID:      config.LocalID,
			Address: transport.LocalAddr(),
		}},
	}
	r.BootstrapCluster(raftConfig)
	log.Printf("[%s] Bootstrapped Raft cluster at %s", shardID, addr)

	return r, kvStore
}
