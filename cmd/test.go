package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"omnidict/client"
	// "omnidict/proto"
	"strconv"
	"context"
)

var testCmd = &cobra.Command{
	Use:   "test [step] [message]",
	Short: "Test gRPC connection and integration",
	Long:  "Test function to verify gRPC connectivity and track integration progress",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Default values
		step := int32(1)
		message := "integration test"

		// Parse step if provided
		if len(args) >= 1 {
			if stepStr, err := strconv.Atoi(args[0]); err == nil {
				step = int32(stepStr)
			}
		}

		// Use custom message if provided
		if len(args) >= 2 {
			message = args[1]
		}

		// 🟢 MOCK version (for now)
		// fmt.Printf("✅ [MOCK] Test Step %d: %s - Connection OK\n", step, message)

		// 📡 Real gRPC version (uncomment this when gRPC server is ready)
		grpcClient, err := client.InitGRPCClient()
		if err != nil {
			fmt.Printf("❌ Failed to connect to gRPC server: %v\n", err)
			return
		}
		defer grpcClient.Close()

		// Create the proper TestRequest
		req := &proto.TestRequest{
			Message: message,
			Step:    step,
		}

		// Call the Test method with proper context
		ctx := context.Background()
		resp, err := grpcClient.Test(ctx, req)
		if err != nil {
			fmt.Printf("❌ [CLIENT] gRPC call failed: %v\n", err)
			return
		}

		fmt.Printf("✅ Test Step %d: %s\n", resp.Step, resp.Message)
		fmt.Printf("✅ Server Status: %s\n", resp.ServerStatus)
		fmt.Printf("✅ Integration milestone reached!\n")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}