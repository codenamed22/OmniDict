package cmd

import (
	"fmt"
	"log"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var ttlCmd = &cobra.Command{
	Use:   "ttl [key]",
	Short: "Get time to live for a key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.TTL(args[0])
		if err != nil {
			return fmt.Errorf("ttl failed: %w", err)
		}

		switch resp.Ttl {
		case -2:
			log.Println("Key does not exist")
		case -1:
			log.Println("Key has no expiration")
		default:
			log.Printf("TTL: %d seconds", resp.Ttl)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ttlCmd)
}
