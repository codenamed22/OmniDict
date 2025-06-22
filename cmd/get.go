package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Retrieve the value for a given key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// MOCK Output – Replace with actual gRPC call later
		fmt.Printf("[MOCK] Value for key '%s' is: 'example_value'\n", key)

		// Example for gRPC (to be added later)
		// resp, err := grpcClient.Get(ctx, &pb.GetRequest{Key: key})
		// if err != nil {
		//     log.Fatalf("Error getting key: %v", err)
		// }
		// fmt.Println("Value:", resp.Value)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
