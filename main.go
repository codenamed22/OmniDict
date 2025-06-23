package main

import (
	"omnidict/client"
	"omnidict/cmd"
)

func main() {
	client.InitGRPCClient()
	defer client.Conn.Close()
	cmd.Execute()
}