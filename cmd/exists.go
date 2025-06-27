package cmd

import (
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var existsCmd = &cobra.Command{
	Use:   "exists [key]",
	Short: "Check if key exists",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Exists(args[0])
		if err != nil {
			log.Fatalf("Exists failed: %v", err)
		}
		log.Printf("Exists: %v", resp.Exists)
	},
}

func init() {
	rootCmd.AddCommand(existsCmd)
}
