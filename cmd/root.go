package cmd

import (
	"fmt"
	"omnidict/client"
	"omnidict/proto"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "omnidict",
	Short: "Distributed key-value store CLI",
}

var putCmd = &cobra.Command{
	Use:   "put [key] [value]",
	Short: "Store a key-value pair",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Put(args[0], []byte(args[1]))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Success: %v\n", resp.Success)
	},
}

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Retrieve a value by key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Get(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Value: %s\n", string(resp.Value))
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a key-value pair",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.Delete(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Deleted: %v\n", resp.Success)
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(deleteCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
