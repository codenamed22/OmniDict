package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"omnidict/client"
	"omnidict/cluster"
	"omnidict/cmd"
	pb_kv "omnidict/proto/kv"
	pb_ring "omnidict/proto/ring"
	"omnidict/ring"
	"omnidict/server"

	"google.golang.org/grpc"
)

// const numShards = 3

func main() {
	var (
		port       = flag.String("port", "8080", "Server port for gRPC")
		raftPort   = flag.String("raft_port", "8081", "Base Raft port for shards")
		serverAddr = flag.String("addr", "localhost:8080", "Server address for CLI")
		nodeID     = flag.String("id", "node1", "Raft node ID")
		dataDir    = flag.String("data", "/tmp/raft", "Raft data directory")
		joinAddr   = flag.String("join", "", "Existing node address to join cluster")
	)
	flag.Parse()

	if err := run(*port, *raftPort, *serverAddr, *nodeID, *dataDir, *joinAddr); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run(port, raftPort, serverAddr, nodeID, dataDir, joinAddr string) error {
	if joinAddr != "" {
		if err := joinCluster(joinAddr, nodeID, raftPort); err != nil {
			return fmt.Errorf("failed to join cluster: %w", err)
		}
	}

	if len(os.Args) > 1 && os.Args[1] == "server" {
		return runServer(port, raftPort, nodeID, dataDir)
	} else {
		return runClient(serverAddr)
	}
}

func joinCluster(joinAddr, nodeID, raftPort string) error {
	conn, err := grpc.Dial(joinAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to join node: %w", err)
	}
	defer conn.Close()

	client := pb_kv.NewOmnidictServiceClient(conn)
	_, err = client.JoinCluster(context.Background(), &pb_kv.JoinRequest{
		NodeId:      nodeID,
		RaftAddress: net.JoinHostPort("0.0.0.0", raftPort),
	})
	if err != nil {
		return fmt.Errorf("failed to join cluster: %w", err)
	}

	log.Printf("Node %s joined cluster via %s", nodeID, joinAddr)
	return nil
}

func runServer(port, raftPort, nodeID, dataDir string) error {
	// Initialize shard manager
	shardMgr := cluster.NewShardManager(nodeID, dataDir, raftPort)
	shardRafts, shardStores, err := shardMgr.InitializeShards()
	if err != nil {
		return fmt.Errorf("failed to initialize shards: %w", err)
	}

	// Initialize hash ring
	hashRing := ring.NewHashRing(10)
	for shardID := range shardRafts {
		hashRing.AddNode(shardID)
		log.Printf("Added shard %s to hash ring", shardID)
	}

	// Start TTL cleaners
	for shardID, store := range shardStores {
		store.StartTTLCleaner(5 * time.Minute)
		log.Printf("Started TTL cleaner for shard %s", shardID)
	}

	// Create ring server
	ringServer := ring.NewRingServer(
		10,          // virtual nodes
		shardStores, // map of shard stores
		net.JoinHostPort("localhost", port),
	)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	grpcServer := grpc.NewServer()

	// Register services
	omniServer := server.NewOmnidictServerWithShards(
		shardStores,
		shardRafts,
		net.JoinHostPort("localhost", port),
	)
	pb_kv.RegisterOmnidictServiceServer(grpcServer, omniServer)
	pb_ring.RegisterRingServiceServer(grpcServer, ringServer)

	log.Printf("Server started at %s (gRPC), Raft base port %s", port, raftPort)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("gRPC serve failed: %w", err)
	}

	return nil
}

func runClient(serverAddr string) error {
	if err := client.InitGRPCClient(serverAddr); err != nil {
		return fmt.Errorf("failed to initialize gRPC client: %w", err)
	}
	defer client.Conn.Close()

	cmd.Execute()
	return nil
}
