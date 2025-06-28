package main

import (
	"flag"
	"log"
	"net"
	"os"
	"time"

	"omnidict/client"
	"omnidict/cmd"
	pb_kv "omnidict/proto/kv"
	pb_ring "omnidict/proto/ring"
	"omnidict/raftstore"
	"omnidict/ring"
	"omnidict/server"
	"omnidict/store"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
)

/*
Raft uses raftPort 8081
gRPC uses port 8080
You can use your own ports by passing them as:
  ./omnidict server --port <> --raft_port <>
*/

func main() {
	var (
		port     = flag.String("port", "8080", "Server port for gRPC")
		raftPort = flag.String("raft_port", "8081", "Raft internal port")
		// Changed serverAddr for CLI to work better in Docker (no 'localhost' inside container)
		serverAddr = flag.String("addr", ":8080", "Server address for CLI")
		nodeID     = flag.String("id", "node1", "Raft node ID")
		dataDir    = flag.String("data", "/tmp/raft", "Raft data directory")
	)
	flag.Parse()

	// If starting in server mode
	if len(os.Args) > 1 && os.Args[1] == "server" {
		config := raft.DefaultConfig()
		config.LocalID = raft.ServerID(*nodeID)
		os.MkdirAll(*dataDir, 0700)

		logStore, err := raftboltdb.NewBoltStore(*dataDir + "/raft-log.bolt")
		if err != nil {
			log.Fatalf("failed to create log store: %v", err)
		}
		stableStore, err := raftboltdb.NewBoltStore(*dataDir + "/raft-stable.bolt")
		if err != nil {
			log.Fatalf("failed to create stable store: %v", err)
		}

		snapshotStore, err := raft.NewFileSnapshotStore(*dataDir, 1, os.Stderr)
		if err != nil {
			log.Fatalf("failed to create snapshot store: %v", err)
		}

		// CHANGED: Use 0.0.0.0 to allow Docker containers to communicate across network
		addr := "0.0.0.0:" + *raftPort
		transport, err := raft.NewTCPTransport(addr, nil, 3, 10*time.Second, os.Stderr)
		if err != nil {
			log.Fatalf("failed to create transport: %v", err)
		}

		// Create the store
		store := store.NewStore()

		// Pass store to Raft FSM
		fsm := raftstore.NewFSM(store)

		r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
		if err != nil {
			log.Fatalf("failed to start Raft node: %v", err)
		}

		raftConfig := raft.Configuration{
			Servers: []raft.Server{{
				ID:      config.LocalID,
				Address: transport.LocalAddr(),
			}},
		}
		r.BootstrapCluster(raftConfig)

		// Start TTL cleaner in background
		store.StartTTLCleaner(5 * time.Minute)

		// Start gRPC server
		lis, err := net.Listen("tcp", ":"+*port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()

		// Register KV RPC service
		omniServer := server.NewOmnidictServerWithStorage(store)
		pb_kv.RegisterOmnidictServiceServer(grpcServer, omniServer)

		// Register Ring RPC service
		ringServer := ring.NewRingServer(10, 3, store, *port)
		pb_ring.RegisterRingServiceServer(grpcServer, ringServer)

		log.Printf("Omnidict server started with Raft at %s (gRPC) and %s (Raft)", *port, *raftPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve failed: %v", err)
		}
	} else {
		// CLI Client logic
		client.InitGRPCClient(*serverAddr)
		defer client.Conn.Close()
		cmd.Execute()
	}
}
