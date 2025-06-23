package cmd

import (
	"context"
	"log"

	"omnidict/client"
	"omnidict/proto"

	"github.com/spf13/cobra"
)

var existsCmd = &cobra.Command{
	Use:   "exists [key]",
	Short: "Check if key exists",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		req := &proto.ExistsRequest{Key: args[0]}
		resp, err := client.Client.Exists(context.Background(), req)
		if err != nil {
			log.Fatalf("Exists failed: %v", err)
		}
		log.Printf("Exists: %v", resp.Exists)
	},
}

func init() {
	rootCmd.AddCommand(existsCmd)
}
