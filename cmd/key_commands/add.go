package key_commands

import (
	"phantom-flux/pkg"
	"phantom-flux/pkg/domain"

	"github.com/spf13/cobra"
)

var KeyAddCmd = &cobra.Command{
	Use:   "add-key",
	Short: "Add SOPS keys to a file",
	Long: `Add SOPS keys to a file. You can specify the Kubernetes context, namespace, secret name, and key name.
			
You can also use the current context with the --from-current-context flag.

Either --from-current-context or --context <ctx-name> must be specified`,

	Run: func(cmd *cobra.Command, args []string) {
		pkg.ExecuteCobraCommand(domain.KeyAdd, cmd, args)
	},
}

func init() {
	pkg.InitCobraCommand(domain.KeyAdd, KeyAddCmd)
}
