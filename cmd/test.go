package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"context"
	"omnidict/client"
	pb "omnidict/proto"
	"strconv"
)

var testCmd = &cobra.Command{
	Use:   "test [step] [message]",
	Short: "Test gRPC connection and integration",
	Long:  "Test function to verify gRPC connectivity and track integration progress",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Default values
		step := int32(1)
		message := "Integration test"

		// Parse step if provided
		if len(args) >= 1 {
			if parsedStep, err := strconv.Atoi(args[0]); err == nil {
				step = int32(parsedStep)
			}
		}

		// Use custom message if provided
		if len(args) >= 2 {
			message = args[1]
		}

		// ✅ MOCK version (for now)
		// fmt.Printf("🧪 [MOCK] Test Step %d: %s - Connection OK\n", step, message)

		// 🔌 Real gRPC version (uncomment this when gRPC server is ready)
		
		resp, err := client.TestConnection(message, step)
		if err != nil {
			fmt.Printf("❌ Test failed at step %d: %v\n", step, err)
			return
		}
		
		fmt.Printf("🧪 Test Step %d: %s\n", resp.Step, resp.Message)
		fmt.Printf("✅ Server Status: %s\n", resp.ServerStatus)
		fmt.Printf("🎯 Integration milestone reached!\n")
		
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}