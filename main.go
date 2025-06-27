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
	"omnidict/ring"
	pb_kv "omnidict/proto/kv"
	pb_ring "omnidict/proto/ring"
	
	"google.golang.org/grpc"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

/*
Raft uses rafrPort 8081
gRPC uses port 8080
u can use your own ports by passing them as
./omnidict server --port <> --raft_port <>
*/

func main() {
	var (
		port       = flag.String("port", "8080", "Server port for gRPC")
		raftPort   = flag.String("raft_port", "8081", "Raft internal port")
		serverAddr = flag.String("addr", "localhost:8080", "Server address for CLI")
		nodeID     = flag.String("id", "node1", "Raft node ID")
		dataDir    = flag.String("data", "/tmp/raft", "Raft data directory")
	)
	flag.Parse()

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

		// Use raftPort for Raft transport
		addr := "127.0.0.1:" + *raftPort
		transport, err := raft.NewTCPTransport(addr, nil, 3, 10*time.Second, os.Stderr)
		if err != nil {
			log.Fatalf("failed to create transport: %v", err)
		}

		// Create store
		store := store.NewStore()
		
		// Create FSM with store reference
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

		// Start TTL cleaner
		store.StartTTLCleaner(5 * time.Minute)

		// Start gRPC server
		lis, err := net.Listen("tcp", ":"+*port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		
		// Register KV service
		omniServer := server.NewOmnidictServerWithStorage(store)
		pb_kv.RegisterOmnidictServiceServer(grpcServer, omniServer)
		
		// Register Ring service
		ringServer := ring.NewRingServer(10, 3, store, *port)
		pb_ring.RegisterRingServiceServer(grpcServer, ringServer)

		log.Printf("Omnidict server started with Raft at %s (gRPC) and %s (Raft)", *port, *raftPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve failed: %v", err)
		}
	} else {
		client.InitGRPCClient(*serverAddr)
		defer client.Conn.Close()
		cmd.Execute()
	}
}
