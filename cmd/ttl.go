package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"omnidict/raftstore" // ✅ Updated to use raftstore directly
)

var ttlCmd = &cobra.Command{
	Use:   "ttl [key]",
	Short: "Get time to live for a key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// ✅ Access FSM via Raft node
		node := raftnode.GetNode()
		if node == nil {
			log.Fatal("No Raft node initialized")
		}

		fsm := node.GetFSM()
		if fsm == nil {
			log.Fatal("FSM not available")
		}

		// ✅ Get TTL for the key
		ttl, err := fsm.TTL(key)
		if err != nil {
			log.Fatalf("TTL failed: %v", err)
		}

		// ✅ Print TTL based on common conventions
		switch ttl {
		case -2:
			log.Println("Key does not exist")
		case -1:
			log.Println("Key has no expiration")
		default:
			log.Printf("TTL: %d seconds", ttl)
		}
	},
}

func init() {
	rootCmd.AddCommand(ttlCmd)
}