package cmd

import (
	"fmt"
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Retrieve a value by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.Get(args[0])
		if err != nil {
			return fmt.Errorf("get failed: %w", err)
		}
		if resp.Found {
			log.Printf("Value: %s", resp.Value)
		} else {
			log.Println("Key not found")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
