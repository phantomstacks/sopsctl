package secret_commands

import (
	"phantom-flux/pkg"
	"phantom-flux/pkg/domain"

	"github.com/spf13/cobra"
)

var SecretCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new SOPS-encrypted secret file",
	Long: `Create a new SOPS-encrypted secret file using the cluster's public AGE key.
The created file will be encrypted and ready to store secrets securely.

Example:
  phantom-flux secret create secrets.yaml --cluster=production`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		pkg.ExecuteCobraCommand(domain.SecretCreate, cmd, args)
	},
}

func init() {
	pkg.InitCobraCommand(domain.SecretCreate, SecretEditCmd)
}
