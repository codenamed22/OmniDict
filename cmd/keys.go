package cmd

import (
	"context"
	"log"
	"strings"

	"omnidict/client"
	"omnidict/proto"

	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys [pattern]",
	Short: "List keys matching pattern",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}
		req := &proto.KeysRequest{Pattern: pattern}
		resp, err := client.Client.Keys(context.Background(), req)
		if err != nil {
			log.Fatalf("Keys failed: %v", err)
		}
		log.Printf("Keys: %s", strings.Join(resp.Keys, ", "))
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
}
