package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ttlCmd = &cobra.Command{
	Use:   "ttl <key>",
	Short: "Show TTL of a key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// 🔄 MOCK
		fmt.Printf("[MOCK] TTL for key '%s': 300 seconds\n", key)

		// 🔌 grpcClient.TTL(ctx, &pb.TTLRequest{Key: key})
	},
}

func init() {
	rootCmd.AddCommand(ttlCmd)
}
