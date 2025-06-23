package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	// 🧠 Uncomment when using real gRPC
	"context"
	"omnidict/client"
	pb "omnidict/proto"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// ✅ MOCK version (for now)
		// fmt.Printf("[MOCK] Deleted key '%s'\n", key)

		
		// 🔌 Real gRPC version (uncomment this when gRPC server is ready)

		_, err := client.GrpcClient.Delete(context.Background(), &pb.DeleteRequest{Key: key})
		if err != nil {
			fmt.Printf("❌ Failed to delete key '%s': %v\n", key, err)
			return
		}
		fmt.Printf("✅ Deleted key '%s'\n", key)
		
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}