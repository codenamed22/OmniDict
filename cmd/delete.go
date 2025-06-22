package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// 🔄 MOCK
		fmt.Printf("[MOCK] Deleted key '%s'\n", key)

		// 🔌 grpcClient.Delete(ctx, &pb.DeleteRequest{Key: key})
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}