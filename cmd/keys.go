package cmd

import (
	"log"
	"strings"

	"omnidict/client"

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
		resp, err := client.Keys(pattern)
		if err != nil {
			log.Fatalf("Keys failed: %v", err)
		}
		log.Printf("Keys: %s", strings.Join(resp.Keys, ", "))
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
}
