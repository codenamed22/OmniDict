package main

import (
	"context"
	"flag"
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

const numShards = 3

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

	if *joinAddr != "" {
		conn, err := grpc.Dial(*joinAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("failed to connect to join node: %v", err)
		}
		defer conn.Close()

		client := pb_kv.NewOmnidictServiceClient(conn)
		_, err = client.JoinCluster(context.Background(), &pb_kv.JoinRequest{
			NodeId:      *nodeID,
			RaftAddress: net.JoinHostPort("0.0.0.0", *raftPort),
		})
		if err != nil {
			log.Fatalf("failed to join cluster: %v", err)
		}
		log.Printf("Node %s joined cluster via %s", *nodeID, *joinAddr)
	}

	if len(os.Args) > 1 && os.Args[1] == "server" {
		// Initialize shard manager
		shardMgr := cluster.NewShardManager(*nodeID, *dataDir, *raftPort)
		shardRafts, shardStores := shardMgr.InitializeShards()

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
			net.JoinHostPort("localhost", *port),
		)

		// Start gRPC server
		lis, err := net.Listen("tcp", ":"+*port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()

		// Register services
		omniServer := server.NewOmnidictServerWithShards(
			shardStores,
			shardRafts,
			net.JoinHostPort("localhost", *port),
		)
		pb_kv.RegisterOmnidictServiceServer(grpcServer, omniServer)
		pb_ring.RegisterRingServiceServer(grpcServer, ringServer)

		log.Printf("Server started at %s (gRPC), Raft base port %s", *port, *raftPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve failed: %v", err)
		}
	} else {
		client.InitGRPCClient(*serverAddr)
		defer client.Conn.Close()
		cmd.Execute()
	}
}
