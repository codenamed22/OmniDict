package main

import (
	"flag"
	"log"
	"net"
	"os"

	"omnidict/cmd"
	"omnidict/client"
	"omnidict/server"
	"omnidict/store"
	"time"

	"google.golang.org/grpc"
	pb_kv "omnidict/proto/kv"
)

func main() {
	var (
		port       = flag.String("port", "8080", "Server port")
		serverAddr = flag.String("addr", "localhost:8080", "Server address for client")
	)
	flag.Parse()

	if len(os.Args) > 1 && os.Args[1] == "server" {
		// Initialize storage
		storage := store.NewStore()
		storage.StartTTLCleaner(5 * time.Minute)
		
		// Create gRPC server
		lis, err := net.Listen("tcp", ":"+*port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		
		// Create and register OmnidictService
		omnidictServer := server.NewOmnidictServerWithStorage(storage) // No arguments needed
		pb_kv.RegisterOmnidictServiceServer(grpcServer, omnidictServer)
		
		log.Printf("Server started on port %s with storage initialized", *port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve failed: %v", err)
		}
	} else {
		// Client mode
		client.InitGRPCClient(*serverAddr)
		defer client.Conn.Close()
		cmd.Execute()
	}
}