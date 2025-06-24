package cmd

import (
	"context"
	"log"

	"omnidict/client"
	"omnidict/proto/kv"

	"github.com/spf13/cobra"
)

var flushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Clear all keys",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		req := &proto.FlushRequest{}
		resp, err := client.Client.Flush(context.Background(), req)
		if err != nil {
			log.Fatalf("Flush failed: %v", err)
		}
		log.Printf("Flush: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(flushCmd)
}
