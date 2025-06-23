package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	// 🧠 Uncomment these when gRPC backend is ready
	// "context"
	// "omnidict/client"
	// pb "omnidict/kvstore/proto"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "List all keys",
	Run: func(cmd *cobra.Command, args []string) {
		// ✅ MOCK version (for now)
		fmt.Println("[MOCK] All keys: user1, session, token")

		/*
		🔌 Real gRPC version (uncomment when ready)

		resp, err := client.GrpcClient.Keys(context.Background(), &pb.KeysRequest{})
		if err != nil {
			fmt.Printf("❌ Failed to list keys: %v\n", err)
			return
		}

		fmt.Print("✅ All keys: ")
		for _, key := range resp.Keys {
			fmt.Print(key, " ")
		}
		fmt.Println()
		*/
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
}