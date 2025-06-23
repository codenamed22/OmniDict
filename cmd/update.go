package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	// 🧠 Uncomment when enabling gRPC
	// "context"
	// "omnidict/client"
	// pb "omnidict/kvstore/proto"
)

var updateCmd = &cobra.Command{
	Use:   "update <key> <value>",
	Short: "Update the value for an existing key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		newVal := args[1]

		// ✅ MOCK version (for now)
		// Assume key exists and simulate update
		fmt.Printf("[MOCK] Updated key '%s' with new value '%s'\n", key, newVal)

		/*
			🔌 Real gRPC version (uncomment when backend is ready)

			_, err := client.GrpcClient.Update(context.Background(), &pb.UpdateRequest{
				Key:   key,
				Value: newVal,
			})
			if err != nil {
				fmt.Printf("❌ Failed to update key '%s': %v\n", key, err)
				return
			}
			fmt.Printf("✅ Updated key '%s' with new value '%s'\n", key, newVal)
		*/
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
