package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"omnidict/raftstore" // ✅ Using raft-aware logic
)

var updateCmd = &cobra.Command{
	Use:   "update [key] [value]",
	Short: "Update an existing key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := []byte(args[1])

		node := raftnode.GetNode()
		if node == nil {
			log.Fatal("No Raft node initialized")
		}

		err := node.ApplyUpdate(key, value)
		if err != nil {
			log.Fatalf("Update failed: %v", err)
		}

		log.Printf("Update successful: %s", key)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}