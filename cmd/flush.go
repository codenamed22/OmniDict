package cmd

import (
	"fmt"
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var flushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Clear all keys",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.Flush()
		if err != nil {
			return fmt.Errorf("flush failed: %w", err)
		}
		log.Printf("Flush: %v", resp)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(flushCmd)
}
