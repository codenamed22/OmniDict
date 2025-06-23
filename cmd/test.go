package cmd

import (
	"context"
	"log"

	"omnidict/client"
	"omnidict/proto"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [message]",
	Short: "Test server connectivity",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Use a simple Put/Get test instead of non-existent TestRequest
		req := &proto.PutRequest{
			Key:   "test_key",
			Value: args[0],
		}
		resp, err := client.Client.Put(context.Background(), req)
		if err != nil {
			log.Fatalf("Test failed: %v", err)
		}
		log.Printf("Test result: %v", resp)
		
		// Test retrieval
		getReq := &proto.GetRequest{Key: "test_key"}
		getResp, err := client.Client.Get(context.Background(), getReq)
		if err != nil {
			log.Fatalf("Test get failed: %v", err)
		}
		log.Printf("Retrieved: %v", getResp)
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
