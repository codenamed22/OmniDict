package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	// 🧠 Uncomment when enabling gRPC
	"context"
	"omnidict/client"
	pb "omnidict/proto"
)

var ttlCmd = &cobra.Command{
	Use:   "ttl <key>",
	Short: "Show TTL of a key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// ✅ MOCK version (for now)
		// fmt.Printf("[MOCK] TTL for key '%s': 300 seconds\n", key)

		
		// 🔌 Real gRPC version (uncomment when backend is ready)

		resp, err := client.GrpcClient.Ttl(context.Background(), &pb.TtlRequest{Key: key})
		if err != nil {
			fmt.Printf("❌ Failed to fetch TTL for key '%s': %v\n", key, err)
			return
		}
		fmt.Printf("✅ TTL for key '%s': %d seconds\n", key, resp.Ttl)
		
	},
}

func init() {
	rootCmd.AddCommand(ttlCmd)
}