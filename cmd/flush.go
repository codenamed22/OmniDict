package cmd

import (
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var flushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Clear all keys",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Flush()
		if err != nil {
			log.Fatalf("Flush failed: %v", err)
		}
		log.Printf("Flush: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(flushCmd)
}
