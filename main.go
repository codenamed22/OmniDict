// use - go build -o omnidict main.go - to generate the user interactive binary file

package main

import (
	"omnidict/cmd"
	"omnidict/client"
	"omnidict/server"
	"omnidict/ring"
	"omnidict/store"
	"os"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		// Initialize storage and start server
		storage := store.NewStore()
		storage.StartTTLCleaner(5 * time.Minute)
		server.StartServer("50051", 200, storage)
	} else {
		// Client mode
		client.InitGRPCClient()
		defer client.Conn.Close()
		cmd.Execute()
	}
}
