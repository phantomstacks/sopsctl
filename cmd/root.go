// Package cmd /*
package cmd

import (
	"os"
	"phantom-flux/cmd/key_commands"
	"phantom-flux/cmd/secret_commands"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sopsctl",
	Short: "Secure configuration management with age encryption and SOPS",
	Long: `Phantom Flux is a CLI tool for managing encrypted configurations using age keys and SOPS in a Kubernetes cluster.

Features:
  - Age key management with secure storage in ~/.pFlux
  - SOPS integration for encrypted YAML/JSON files
  - Interactive text editor (use 'sopsctl edit' command)
  - Easy key generation and import

Get started:
  sopctl key init      # Initialize your age key
  sopctl key show      # Display your key information
  sopctl --help        # Show all available commands`,
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

	rootCmd.AddCommand(secret_commands.SecretDecryptCmd)
	rootCmd.AddCommand(secret_commands.SecretEditCmd)
	rootCmd.AddCommand(secret_commands.SecretCreateCmd)

	rootCmd.AddCommand(key_commands.KeyAddCmd)
	rootCmd.AddCommand(key_commands.KeyListCmd)
	rootCmd.AddCommand(key_commands.RemoveCmd)
	rootCmd.AddCommand(key_commands.KeyStorageModeCmd)
}
