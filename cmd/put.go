package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	// 🧠 Uncomment when enabling gRPC
	"context"
	"omnidict/client"
	pb "omnidict/proto"
)

var putCmd = &cobra.Command{
	Use:   "put <key> <value>",
	Short: "Store a key-value pair",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		// ✅ MOCK version (for now)
		// fmt.Printf("[MOCK] Stored key '%s' with value '%s'\n", key, value)

		
		// 🔌 Real gRPC version (uncomment when backend is live)

		_, err := client.GrpcClient.Put(context.Background(), &pb.PutRequest{
			Key:   key,
			Value: value,
		})
		if err != nil {
			fmt.Printf("❌ Failed to store key '%s': %v\n", key, err)
			return
		}
		fmt.Printf("✅ Stored key '%s' with value '%s'\n", key, value)
		
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}