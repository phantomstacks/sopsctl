package secret_commands

import (
	"phantom-flux/pkg"
	"phantom-flux/pkg/domain"

	"github.com/spf13/cobra"
)

var SecretEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a SOPS-encrypted secret file in your default editor",
	Long: `Edit a SOPS-encrypted secret file securely using your default editor.

This command provides a secure workflow for editing encrypted secrets:
1. Decrypts the file using the cluster's private AGE key
2. Opens the decrypted content in your system's default editor
3. After you save and close the editor, re-encrypts the content with the cluster's public key
4. Atomically writes the encrypted content back to the original file

The original encrypted file is never exposed in plain text on disk except in a
temporary file during editing. The command ensures data integrity through atomic
file operations.

Example:
  phantom-flux secret edit secrets.yaml --cluster=production`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.ExecuteCobraCommand(domain.SecretEdit, cmd, args)
	},
}

func init() {
	pkg.InitCobraCommand(domain.SecretEdit, SecretEditCmd)
}
