package cmd

import (
	"context"
	"log"

	"omnidict/client"
	"omnidict/proto"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a key-value pair",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		req := &proto.DeleteRequest{Key: args[0]}
		resp, err := client.Client.Delete(context.Background(), req)
		if err != nil {
			log.Fatalf("Delete failed: %v", err)
		}
		log.Printf("Delete: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
