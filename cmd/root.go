// Package cmd /*
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pflux",
	Short: "Secure configuration management with age encryption and SOPS",
	Long: `Phantom Flux is a CLI tool for managing encrypted configurations using age keys and SOPS in a Kubernetes cluster.

Features:
  - Age key management with secure storage in ~/.pFlux
  - SOPS integration for encrypted YAML/JSON files
  - Interactive text editor (use 'pflux editor' command)
  - Easy key generation and import

Get started:
  pflux key init      # Initialize your age key
  pflux key show      # Display your key information
  pflux --help        # Show all available commands`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("cluster", "c", "", "Kubernetes cluster context to use")
}
