package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"omnidict/raftstore" // ✅ Changed to use raftstore
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Retrieve a value by key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// ✅ Access the Raft FSM
		node := raftnode.GetNode()
		if node == nil {
			log.Fatal("No Raft node initialized")
		}

		fsm := node.GetFSM()
		if fsm == nil {
			log.Fatal("FSM not available")
		}

		// ✅ Get value from FSM
		value, found := fsm.Get(key)
		if found {
			log.Printf("Value: %s", value)
		} else {
			log.Println("Key not found")
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}