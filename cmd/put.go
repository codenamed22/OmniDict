package cmd

import (
	"fmt"
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put [key] [value]",
	Short: "Store a key-value pair",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.Put(args[0], []byte(args[1]))
		if err != nil {
			return fmt.Errorf("put failed: %w", err)
		}
		log.Printf("Put: %v", resp)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}
