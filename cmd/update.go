package cmd

import (
	"fmt"
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [key] [value]",
	Short: "Update an existing key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.Update(args[0], []byte(args[1]))
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		log.Printf("Update: %v", resp)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
