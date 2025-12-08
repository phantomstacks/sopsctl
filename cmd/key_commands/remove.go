package key_commands

import (
	"sopsctl/pkg"
	"sopsctl/pkg/domain"

	"github.com/spf13/cobra"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove-key",
	Short: "Remove SOPS keys",
	Long:  `Remove SOPS keys from local storage.`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.ExecuteCobraCommand(domain.KeyRemove, cmd, args)
	},
}

func init() {
	pkg.InitCobraCommand(domain.KeyRemove, RemoveCmd)
}
