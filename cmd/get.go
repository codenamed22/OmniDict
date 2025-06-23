package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	// 🧠 Uncomment when enabling gRPC
	"context"
	"omnidict/client"
	pb "omnidict/proto"
)

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Retrieve the value for a given key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// ✅ MOCK version (for now)
		// fmt.Printf("[MOCK] Value for key '%s' is: 'example_value'\n", key)

		
		// 🔌 Real gRPC version (uncomment this when gRPC is ready)

		resp, err := client.GrpcClient.Get(context.Background(), &pb.GetRequest{Key: key})
		if err != nil {
			fmt.Printf("❌ Failed to get key '%s': %v\n", key, err)
			return
		}
		fmt.Printf("✅ Value for key '%s' is: '%s'\n", key, resp.Value)
		
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}