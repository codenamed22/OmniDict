package cmd

import (
	"encoding/json"
	"log"

	"omnidict/raftstore" // ✅ You *are* using raftstore, now using properly
	"github.com/spf13/cobra"
)

// ✅ RaftNode will be injected from main.go or setup code
var RaftNode *raftnode.Node

// Struct representing a delete command for FSM
type DeleteCommand struct {
	Op  string `json:"op"`  // "delete"
	Key string `json:"key"` // key to delete
}

var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a key-value pair",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if RaftNode == nil {
			log.Fatal("Raft node not initialized")
		}

		key := args[0]

		// 🚀 Create Raft-compatible command object
		delCmd := DeleteCommand{
			Op:  "delete",
			Key: key,
		}

		// 🧠 Serialize the command to JSON
		data, err := json.Marshal(delCmd)
		if err != nil {
			log.Fatalf("Failed to serialize delete command: %v", err)
		}

		// 📤 Propose to Raft node
		err = RaftNode.Propose(data)
		if err != nil {
			log.Fatalf("Raft proposal for delete failed: %v", err)
		}

		log.Printf("✅ Delete command proposed via Raft: key=%s", key)
	},
}

// ✅ This will be called from main.go to inject the raft node
func InjectRaftNode(node *raftnode.Node) {
	RaftNode = node
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
