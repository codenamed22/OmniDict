package cmd

import (
	// 🧠 Uncomment when enabling gRPC
	"omnidict/client"
	
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "omnidict",
	Short: "OmniDict CLI - Distributed KV Store",
}

func Execute() {
	// ✅ Initialize gRPC client connection before any command runs
	client.InitGRPCClient()

	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Add all subcommands here
	// rootCmd.AddCommand(deleteCmd)
	// rootCmd.AddCommand(existsCmd)
	// rootCmd.AddCommand(expireCmd)
	// rootCmd.AddCommand(flushCmd)
	// rootCmd.AddCommand(getCmd)
	// rootCmd.AddCommand(keysCmd)
	// rootCmd.AddCommand(putCmd)
	// rootCmd.AddCommand(ttlCmd)
	// rootCmd.AddCommand(updateCmd)
}