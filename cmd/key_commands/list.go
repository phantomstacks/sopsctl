package key_commands

import (
	"phantom-flux/pkg"
	"phantom-flux/pkg/domain"

	"github.com/spf13/cobra"
)

var KeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SOPS keys",
	Long:  `List all SOPS keys stored locally.`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.ExecuteCobraCommand(domain.KeyList, cmd, args)
	},
}

func init() {
	pkg.InitCobraCommand(domain.KeyList, KeyListCmd)
}
