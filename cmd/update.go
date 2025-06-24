package cmd

import (
	"context"
	"log"

	"omnidict/client"
	"omnidict/proto/kv"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [key] [value]",
	Short: "Update an existing key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		req := &proto.UpdateRequest{
			Key:   args[0],
			Value: args[1],
		}
		resp, err := client.Client.Update(context.Background(), req)  // Changed from client.GrpcClient to client.Client
		if err != nil {
			log.Fatalf("Update failed: %v", err)
		}
		log.Printf("Update: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
