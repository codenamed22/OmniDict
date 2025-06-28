package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"omnidict/raftstore" // ✅ Changed from client to raftstore
)

var putCmd = &cobra.Command{
	Use:   "put [key] [value]",
	Short: "Store a key-value pair",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := []byte(args[1])

		// ✅ Access Raft node
		node := raftnode.GetNode()
		if node == nil {
			log.Fatal("No Raft node initialized")
		}

		// ✅ Apply the put through the Raft log
		err := node.ApplyPut(key, value)
		if err != nil {
			log.Fatalf("Put failed: %v", err)
		}

		log.Printf("Put successful: %s", key)
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}