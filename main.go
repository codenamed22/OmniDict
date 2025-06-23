// use - go build -o omnidict main.go - to generate the user interactive binary file

package main

import (
	"omnidict/cmd"
	"omnidict/client"
	"omnidict/server"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		server.StartServer()
	} else {
		client.InitGRPCClient()
		defer client.Conn.Close()
		cmd.Execute()
	}
}
