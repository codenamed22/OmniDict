package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "omnidict",
	Short: "OmniDict CLI – Distributed KV Store",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
