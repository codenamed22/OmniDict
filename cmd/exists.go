package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"omnidict/raftstore" // ✅ Added for Raft support to access FSM
)

var existsCmd = &cobra.Command{
	Use:   "exists [key]",
	Short: "Check if key exists",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// ✅ Raft support: Access current node's FSM and call Exists
		node := raftnode.GetNode() // 🔄 CHANGE: Assuming you expose a singleton or instance method to get the current node
		if node == nil {
			log.Fatal("No Raft node initialized")
		}

		fsm := node.GetFSM() // 🔄 CHANGE: Get FSM from node
		if fsm == nil {
			log.Fatal("FSM not available")
		}

		exists, err := fsm.Exists(key) // 🔄 CHANGE: Call Exists on FSM
		if err != nil {
			log.Fatalf("Exists check failed: %v", err)
		}

		log.Printf("Exists: %v", exists)
	},
}

func init() {
	rootCmd.AddCommand(existsCmd)
}