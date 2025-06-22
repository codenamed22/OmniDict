package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "List all keys",
	Run: func(cmd *cobra.Command, args []string) {
		// 🔄 MOCK
		fmt.Println("[MOCK] All keys: user1, session, token")

		// 🔌 grpcClient.Keys(ctx, &pb.KeysRequest{})
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
}