package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"omnidict/raftstore" // ✅ Raft-aware package usage
)

var flushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Clear all keys",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// ✅ Get current Raft node
		node := raftnode.GetNode()
		if node == nil {
			log.Fatal("No Raft node initialized")
		}

		fsm := node.GetFSM()
		if fsm == nil {
			log.Fatal("FSM not available")
		}

		// ✅ Call the Flush method on the FSM
		err := fsm.Flush()
		if err != nil {
			log.Fatalf("Flush failed: %v", err)
		}

		log.Println("Flush: all keys cleared successfully")
	},
}

func init() {
	rootCmd.AddCommand(flushCmd)
}