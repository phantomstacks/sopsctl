package secret_commands

import (
	"phantom-flux/pkg"
	"phantom-flux/pkg/domain"

	"github.com/spf13/cobra"
)

var SecretCreateCmd = &cobra.Command{
	Use:   "create NAME [--type=string] [--from-file=[key=]source] [--from-literal=key1=value1] [--from-env-file=source]",
	Short: "Create an encrypted secret from files, directories, or literal values",
	Long: `Create an encrypted Kubernetes secret from a local file, directory, or literal value.
The secret will be encrypted and output as YAML.

When creating a secret based on a file, the key will default to the basename of the file, 
and the value will default to the file content. If the basename is an invalid key or you 
wish to choose your own, you may specify an alternate key.

When creating a secret based on a directory, each file whose basename is a valid key in 
the directory will be packaged into the secret.`,
	Example: `  # Create a new secret named my-secret with keys for each file in folder bar
  phantom-flux secret create my-secret --from-file=path/to/bar

  # Create a new secret named my-secret with specified keys instead of names on disk
  phantom-flux secret create my-secret --from-file=ssh-privatekey=path/to/id_rsa --from-file=ssh-publickey=path/to/id_rsa.pub

  # Create a new secret named my-secret with key1=supersecret and key2=topsecret
  phantom-flux secret create my-secret --from-literal=key1=supersecret --from-literal=key2=topsecret

  # Create a new secret from env files
  phantom-flux secret create my-secret --from-env-file=path/to/foo.env --from-env-file=path/to/bar.env`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.ExecuteCobraCommand(domain.SecretCreate, cmd, args)
	},
}

func init() {
	pkg.InitCobraCommand(domain.SecretCreate, SecretCreateCmd)
}
