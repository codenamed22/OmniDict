package cmd

import (
	"context"
	"log"

	"omnidict/client"
	"omnidict/proto"

	"github.com/spf13/cobra"
)

var ttlCmd = &cobra.Command{
	Use:   "ttl [key]",
	Short: "Get time to live for a key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		req := &proto.TTLRequest{Key: args[0]}  // Changed from pb.TtlRequest to proto.TTLRequest
		resp, err := client.Client.TTL(context.Background(), req)  // Changed from client.GrpcClient to client.Client
		if err != nil {
			log.Fatalf("TTL failed: %v", err)
		}
		
		switch resp.Ttl {
		case -2:
			log.Println("Key does not exist")
		case -1:
			log.Println("Key has no expiration")
		default:
			log.Printf("TTL: %d seconds", resp.Ttl)
		}
	},
}

func init() {
	rootCmd.AddCommand(ttlCmd)
}
