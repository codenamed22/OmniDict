package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <key> <value>",
	Short: "Update the value for an existing key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		newVal := args[1]

		// 🔄 MOCK implementation
		// Assume key exists and simulate update
		fmt.Printf("[MOCK] Updated key '%s' with new value '%s'\n", key, newVal)

		// 🔌 Future:
		// resp, err := grpcClient.Update(ctx, &pb.UpdateRequest{Key: key, Value: newVal})
		// if err != nil {
		//     fmt.Println("Update failed:", err)
		//     return
		// }
		// fmt.Println("Update successful:", resp.Status)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
