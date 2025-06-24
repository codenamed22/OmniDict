package cmd

import (
	"context"
	"log"

	"omnidict/client"
	"/kv"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Retrieve a value by key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		req := &proto.GetRequest{Key: args[0]}
		resp, err := client.Client.Get(context.Background(), req)
		if err != nil {
			log.Fatalf("Get failed: %v", err)
		}
		if resp.Found {
			log.Printf("Value: %s", resp.Value)
		} else {
			log.Println("Key not found")
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
