package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"omnidict/client"
	"omnidict/cluster"
	"omnidict/cmd"
	"omnidict/gossip"
	pb_gossip "omnidict/proto/gossip"
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
		dataDir    = flag.String("data", "../data", "Raft data directory")
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
	// Initialize gossip manager
	gossipMgr := gossip.NewGossipManager(nodeID, []string{})
	gossipMgr.AddMember(nodeID) // Add self to membership
	gossipMgr.Start()

	// Initialize shard manager
	shardMgr := cluster.NewShardManager(nodeID, dataDir, raftPort)
	shardRafts, shardStores, err := shardMgr.InitializeShards()
	if err != nil {
		return fmt.Errorf("failed to initialize shards: %v", err)
	}

	// Create server
	omniServer := server.NewOmnidictServerWithShards(
		shardStores,
		shardRafts,
		fmt.Sprintf(":%s", port),
		gossipMgr,
	)

	// Initialize ring server
	ringServer := ring.NewRingServer(3, shardStores, nodeID)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	defer lis.Close()

	// Register services
	gossipServer := gossip.NewGossipServer(gossipMgr)
	grpcServer := grpc.NewServer()
	pb_kv.RegisterOmnidictServiceServer(grpcServer, omniServer)
	pb_ring.RegisterRingServiceServer(grpcServer, ringServer)
	pb_gossip.RegisterGossipServiceServer(grpcServer, gossipServer)

	log.Printf("Server started on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
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
