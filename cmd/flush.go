package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var flushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Delete all keys (dangerous!)",
	Run: func(cmd *cobra.Command, args []string) {
		// 🔄 MOCK
		fmt.Println("[MOCK] All keys flushed")

		// 🔌 grpcClient.Flush(ctx, &pb.FlushRequest{})
	},
}

func init() {
	rootCmd.AddCommand(flushCmd)
}