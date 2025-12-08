package secret_commands

import (
	"sopsctl/pkg"
	"sopsctl/pkg/domain"

	"github.com/spf13/cobra"
)

var SecretDecryptCmd = &cobra.Command{
	Use:   "decrypt <file>",
	Short: "Decrypt a SOPS-encrypted secret file using AGE encryption",
	Long: `Decrypt a SOPS-encrypted secret file using AGE encryption keys.

This command retrieves the private AGE key for the specified cluster from the key manager
and uses it to decrypt the provided file. The decrypted content is output to stdout.

The file should be encrypted with SOPS using AGE encryption. The command supports
YAML formatted files and requires a valid cluster context to retrieve the correct
decryption key.

Example:
  sopsctl secret decrypt secret.yaml --cluster=production`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pkg.ExecuteCobraCommand(domain.SecretDecrypt, cmd, args)
	},
}

func init() {
}
