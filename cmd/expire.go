package cmd

import (
	"fmt"
	"log"
	"strconv"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var expireCmd = &cobra.Command{
	Use:   "expire [key] [seconds]",
	Short: "Set expiration time for a key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ttl, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid TTL: %w", err)
		}

		resp, err := client.Expire(args[0], ttl)
		if err != nil {
			return fmt.Errorf("expire failed: %w", err)
		}
		log.Printf("Expire: %v", resp)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(expireCmd)
}
