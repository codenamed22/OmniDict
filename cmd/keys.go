package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"

	"omnidict/raftstore" // ✅ Replaced client with raftstore
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

		// ✅ Access the Raft FSM
		node := raftnode.GetNode()
		if node == nil {
			log.Fatal("No Raft node initialized")
		}

		fsm := node.GetFSM()
		if fsm == nil {
			log.Fatal("FSM not available")
		}

		// ✅ Get matching keys
		keys, err := fsm.Keys(pattern)
		if err != nil {
			log.Fatalf("Keys failed: %v", err)
		}

		if len(keys) == 0 {
			log.Println("No keys found")
		} else {
			log.Printf("Keys: %s", strings.Join(keys, ", "))
		}
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
}