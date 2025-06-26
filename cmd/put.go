package cmd

import (
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put [key] [value]",
	Short: "Store a key-value pair",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Put(args[0], []byte(args[1]))
		if err != nil {
			log.Fatalf("Put failed: %v", err)
		}
		log.Printf("Put: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}
