package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "omnidict",
	Short: "Distributed key-value store",
	Long:  "Omnidict - A distributed key-value store using gRPC and consistent hashing",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
