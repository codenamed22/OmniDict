package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	// 🧠 Uncomment these when switching to gRPC
	// "context"
	// "omnidict/client"
	// pb "omnidict/kvstore/proto"
)

var expireCmd = &cobra.Command{
	Use:   "expire <key> <seconds>",
	Short: "Set TTL on a key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		ttl, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid TTL value")
			return
		}

		// ✅ MOCK version (for now)
		fmt.Printf("[MOCK] TTL for key '%s' set to %d seconds\n", key, ttl)

		/*
		🔌 Real gRPC version (uncomment this when gRPC is active)

		_, err = client.GrpcClient.Expire(context.Background(), &pb.ExpireRequest{
			Key: key,
			TTL: int64(ttl),
		})
		if err != nil {
			fmt.Printf("❌ Failed to set TTL for key '%s': %v\n", key, err)
			return
		}
		fmt.Printf("✅ TTL for key '%s' set to %d seconds\n", key, ttl)
		*/
	},
}

func init() {
	rootCmd.AddCommand(expireCmd)
}