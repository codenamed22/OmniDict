package cmd

import (
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [key] [value]",
	Short: "Update an existing key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Update(args[0], []byte(args[1]))
		if err != nil {
			log.Fatalf("Update failed: %v", err)
		}
		log.Printf("Update: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
