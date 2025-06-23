// use - go build -o omnidict main.go - to generate the user interactive binary file

package main

import (
	"omnidict/cmd"
)

func main() {
	client.InitGRPCClient()
	defer client.Conn.Close()
	cmd.Execute()
}