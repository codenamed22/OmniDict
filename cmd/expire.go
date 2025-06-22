package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var expireCmd = &cobra.Command{
	Use:   "expire <key> <seconds>",
	Short: "Set TTL on a key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		ttl, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid TTL value")
			return
		}

		// 🔄 MOCK
		fmt.Printf("[MOCK] TTL for key '%s' set to %d seconds\n", key, ttl)

		// 🔌 grpcClient.Expire(ctx, &pb.ExpireRequest{Key: key, TTL: int64(ttl)})
	},
}

func init() {
	rootCmd.AddCommand(expireCmd)
}
