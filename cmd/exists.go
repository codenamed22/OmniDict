package cmd

import (
	"fmt"
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var existsCmd = &cobra.Command{
	Use:   "exists [key]",
	Short: "Check if key exists",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.Exists(args[0])
		if err != nil {
			return fmt.Errorf("exists failed: %w", err)
		}
		log.Printf("Exists: %v", resp.Exists)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(existsCmd)
}
