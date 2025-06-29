package cmd

import (
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a key-value pair",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Delete(args[0])
		if err != nil {
			log.Fatalf("Delete failed: %v", err)
		}
		log.Printf("Delete: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
