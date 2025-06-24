package cmd

import (
	"context"
	"log"
	"strconv"

	"omnidict/client"
	"omnidict/prot/kv"

	"github.com/spf13/cobra"
)

var expireCmd = &cobra.Command{
	Use:   "expire [key] [seconds]",
	Short: "Set expiration time for a key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ttl, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			log.Fatalf("Invalid TTL: %v", err)
		}
		
		req := &proto.ExpireRequest{
			Key: args[0],
			Ttl: ttl,  // Changed from TtlSeconds to Ttl
		}
		resp, err := client.Client.Expire(context.Background(), req)  // Changed from client.GrpcClient to client.Client
		if err != nil {
			log.Fatalf("Expire failed: %v", err)
		}
		log.Printf("Expire: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(expireCmd)
}
