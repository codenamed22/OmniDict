package cmd

import (
	"context"
	"fmt"
	"strings"

	"omnidict/client"
	pb "omnidict/proto"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a key from the distributed store",
	Long: `Delete removes a key-value pair from the distributed key-value store.
	
The key will be deleted from the appropriate node based on consistent hashing.
If the key doesn't exist, an error will be returned.

Example:
  omnidict delete mykey`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.TrimSpace(args[0])
		
		if key == "" {
			fmt.Println("❌ Error: Key cannot be empty")
			return
		}

		fmt.Printf("🗑️  [DELETE] Attempting to delete key: '%s'\n", key)
		
		// Call the delete function
		if err := deleteKey(key); err != nil {
			fmt.Printf("❌ [DELETE] Operation failed: %v\n", err)
			return
		}
	},
}

// deleteKey performs the actual delete operation
func deleteKey(key string) error {
	fmt.Printf("📡 [CLIENT] Preparing delete request for key: '%s'\n", key)
	fmt.Printf("🔗 [CLIENT] Sending gRPC delete request to server...\n")
	
	// Make the gRPC call to delete the key
	response, err := client.GrpcClient.Delete(context.Background(), &pb.DeleteRequest{
		Key: key,
	})
	
	if err != nil {
		fmt.Printf("❌ [CLIENT] gRPC call failed: %v\n", err)
		return fmt.Errorf("failed to delete key '%s': %v", key, err)
	}
	
	fmt.Printf("✅ [CLIENT] Received response from server\n")
	
	// Check the response
	if response.GetSuccess() {
		fmt.Printf("✅ [SUCCESS] Key '%s' deleted successfully\n", key)
		fmt.Printf("🎯 [SERVER] Delete operation completed\n")
	} else {
		errorMsg := response.GetError()
		if errorMsg == "" {
			errorMsg = "Unknown error occurred"
		}
		fmt.Printf("❌ [FAILED] Could not delete key '%s': %s\n", key, errorMsg)
		return fmt.Errorf("delete operation failed: %s", errorMsg)
	}
	
	return nil
}

func init() {
	// Register the delete command with the root command
	rootCmd.AddCommand(deleteCmd)
}