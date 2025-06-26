package main

import (
	"flag"
	"log"
	"net"
	"os"
	"time"

	"omnidict/cmd"
	"omnidict/client"
	"omnidict/server"
	"omnidict/store"
	"omnidict/raftstore"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
	pb_kv "omnidict/proto/kv"
)

func main() {
	// CLI flags to choose port, address and Raft node ID
	var (
		port       = flag.String("port", "8080", "Server port")
		serverAddr = flag.String("addr", "localhost:8080", "Server address for CLI")
		nodeID     = flag.String("id", "node1", "Raft node ID")
		dataDir    = flag.String("data", "/tmp/raft", "Raft data directory")
	)
	flag.Parse()

	// If run as server -> start Raft + gRPC
	if len(os.Args) > 1 && os.Args[1] == "server" {
		// Create Raft config
		config := raft.DefaultConfig()
		config.LocalID = raft.ServerID(*nodeID)

		// Ensure Raft data folder exists
		os.MkdirAll(*dataDir, 0700)

		// Create BoltDB stores for Raft logs and stable storage
		logStore, err := raftboltdb.NewBoltStore(*dataDir + "/raft-log.bolt")
		if err != nil {
			log.Fatalf("failed to create log store: %v", err)
		}
		stableStore, err := raftboltdb.NewBoltStore(*dataDir + "/raft-stable.bolt")
		if err != nil {
			log.Fatalf("failed to create stable store: %v", err)
		}

		// Snapshot store for Raft state
		snapshotStore, err := raft.NewFileSnapshotStore(*dataDir, 1, os.Stderr)
		if err != nil {
			log.Fatalf("failed to create snapshot store: %v", err)
		}

		// Raft transport (bind to IP:port)
		addr := "127.0.0.1:" + *port
		transport, err := raft.NewTCPTransport(addr, nil, 3, 10*time.Second, os.Stderr)
		if err != nil {
			log.Fatalf("failed to create transport: %v", err)
		}

		// Create FSM and Raft node
		fsm := raftnode.NewFSM()
		r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
		if err != nil {
			log.Fatalf("failed to start Raft node: %v", err)
		}

		// Bootstrap cluster (for single-node MVP only)
		raftConfig := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		r.BootstrapCluster(raftConfig)

		// Set up TTL store
		store := store.NewStoreWithFSM(fsm)
		store.StartTTLCleaner(5 * time.Minute)

		// Start gRPC server
		lis, err := net.Listen("tcp", ":"+*port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		omniServer := server.NewOmnidictServerWithStorage(store)
		pb_kv.RegisterOmnidictServiceServer(grpcServer, omniServer)

		log.Printf("Omnidict server started with Raft at %s", addr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve failed: %v", err)
		}
	} else {
		// CLI Mode: Start client and execute commands
		client.InitGRPCClient(*serverAddr)
		defer client.Conn.Close()
		cmd.Execute()
	}
}