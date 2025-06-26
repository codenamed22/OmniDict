// use - go build -o omnidict main.go - to generate the user interactive binary file

package main

import (
	"omnidict/cmd"
	"omnidict/client"
	"omnidict/ring"
	"omnidict/store"
	"os"
	"time"
	"flag"
	"log"
)

func main() {
	// Define command line flags for server mode
	var (
		port         = flag.String("port", "8080", "Server port")
		virtualNodes = flag.Int("vnodes", 150, "Number of virtual nodes")
		serverAddr   = flag.String("addr", "localhost:8080", "Server address for client")
	)
	flag.Parse()

	if len(os.Args) > 1 && os.Args[1] == "server" {
		// Initialize storage and start server
		storage := store.NewStore()
		storage.StartTTLCleaner(5 * time.Minute)
		
		log.Printf("Starting distributed server on port %s with %d virtual nodes", *port, *virtualNodes)
		ring.StartServer(*port, *virtualNodes, storage)
	} else {
		// Client mode - connect to the server
		client.InitGRPCClient(*serverAddr)
		defer client.Conn.Close()
		cmd.Execute()
	}
}