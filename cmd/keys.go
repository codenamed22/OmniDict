package cmd

import (
	"fmt"
	"log"
	"strings"

	"omnidict/client"

	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys [pattern]",
	Short: "List keys matching pattern",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}
		resp, err := client.Keys(pattern)
		if err != nil {
			return fmt.Errorf("keys failed: %w", err)
		}
		log.Printf("Keys: %s", strings.Join(resp.Keys, ", "))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
}
