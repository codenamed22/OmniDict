package cmd

import (
	"context"
	"log"

	"omnidict/client"
	"omnidict/proto/kv"

	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put [key] [value]",
	Short: "Store a key-value pair",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		req := &proto.PutRequest{
			Key:   args[0],
			Value: args[1],
		}
		resp, err := client.Client.Put(context.Background(), req)
		if err != nil {
			log.Fatalf("Put failed: %v", err)
		}
		log.Printf("Put: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}
