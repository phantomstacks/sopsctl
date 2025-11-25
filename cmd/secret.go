// Package cmd /*
package cmd

import (
	"phantom-flux/cmd/secret_commands"

	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage encrypted secret files with SOPS",
	Long: `Manage your encrypted secret files using SOPS integration.
This command allows you to decrypt and edit secret files securely. 
It uses age keys for encryption and decryption.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(secretCmd)
	secretCmd.AddCommand(secret_commands.SecretDecryptCmd)
	secretCmd.AddCommand(secret_commands.SecretEditCmd)
}
