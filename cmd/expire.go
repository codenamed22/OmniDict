package cmd

import (
	"log"
	"strconv"

	"github.com/spf13/cobra"

	"omnidict/raftstore" // ✅ Changed from client to raftstore
)

var expireCmd = &cobra.Command{
	Use:   "expire [key] [seconds]",
	Short: "Set expiration time for a key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// ✅ Convert TTL from string to int64
		ttl, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			log.Fatalf("Invalid TTL: %v", err)
		}

		// ✅ Access the Raft node's FSM
		node := raftnode.GetNode()
		if node == nil {
			log.Fatal("No Raft node initialized")
		}

		fsm := node.GetFSM()
		if fsm == nil {
			log.Fatal("FSM not available")
		}

		// ✅ Set expiration using Raft FSM
		err = fsm.Expire(key, ttl)
		if err != nil {
			log.Fatalf("Expire failed: %v", err)
		}

		log.Printf("Expire set for key '%s' with TTL %d seconds", key, ttl)
	},
}

func init() {
	rootCmd.AddCommand(expireCmd)
}