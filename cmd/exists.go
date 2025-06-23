package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	// 🧠 Uncomment when enabling gRPC
	"context"
	"omnidict/client"
	pb "omnidict/proto"
)

var existsCmd = &cobra.Command{
	Use:   "exists <key>",
	Short: "Check if a key exists",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// ✅ MOCK version (for now)
		// fmt.Printf("[MOCK] Key '%s' exists: true\n", key)

		
		// 🔌 Real gRPC version (uncomment this when gRPC is active)

		resp, err := client.GrpcClient.Exists(context.Background(), &pb.ExistsRequest{Key: key})
		if err != nil {
			fmt.Printf("❌ Failed to check key '%s': %v\n", key, err)
			return
		}
		fmt.Printf("✅ Key '%s' exists: %v\n", key, resp.Exists)
		
	},
}

func init() {
	rootCmd.AddCommand(existsCmd)
}