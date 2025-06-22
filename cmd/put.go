package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put <key> <value>",
	Short: "Store a key-value pair",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		fmt.Printf("[MOCK] Stored key '%s' with value '%s'\n", key, value)//replace with grpc
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}
