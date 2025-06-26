package cmd

import (
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Retrieve a value by key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get(args[0])
		if err != nil {
			log.Fatalf("Get failed: %v", err)
		}
		if resp.Found {
			log.Printf("Value: %s", resp.Value)
		} else {
			log.Println("Key not found")
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
