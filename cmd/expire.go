package cmd

import (
	"log"
	"strconv"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var expireCmd = &cobra.Command{
	Use:   "expire [key] [seconds]",
	Short: "Set expiration time for a key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ttl, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			log.Fatalf("Invalid TTL: %v", err)
		}
		
		resp, err := client.Expire(args[0], ttl)
		if err != nil {
			log.Fatalf("Expire failed: %v", err)
		}
		log.Printf("Expire: %v", resp)
	},
}

func init() {
	rootCmd.AddCommand(expireCmd)
}
