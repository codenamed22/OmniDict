package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var existsCmd = &cobra.Command{
	Use:   "exists <key>",
	Short: "Check if a key exists",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// 🔄 MOCK
		fmt.Printf("[MOCK] Key '%s' exists: true\n", key)

		// 🔌 grpcClient.Exists(ctx, &pb.ExistsRequest{Key: key})
	},
}

func init() {
	rootCmd.AddCommand(existsCmd)
}