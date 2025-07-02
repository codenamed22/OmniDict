package cmd

import (
	"fmt"
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a key-value pair",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.Delete(args[0])
		if err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}
		log.Printf("Delete: %v", resp)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
