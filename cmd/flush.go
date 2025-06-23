package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	// 🧠 Uncomment when using real gRPC
	"context"
	"omnidict/client"
	pb "omnidict/proto"
)

var flushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Delete all keys (dangerous!)",
	Run: func(cmd *cobra.Command, args []string) {
		// ✅ MOCK version (for now)
		// fmt.Println("[MOCK] All keys flushed")

		
		// 🔌 Real gRPC version (uncomment this when gRPC is active)

		_, err := client.GrpcClient.Flush(context.Background(), &pb.FlushRequest{})
		if err != nil {
			fmt.Printf("❌ Failed to flush keys: %v\n", err)
			return
		}
		fmt.Println("✅ All keys flushed successfully")
		
	},
}

func init() {
	rootCmd.AddCommand(flushCmd)
}