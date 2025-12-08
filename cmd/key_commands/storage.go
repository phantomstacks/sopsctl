package key_commands

import (
	"sopsctl/pkg"
	"sopsctl/pkg/domain"

	"github.com/spf13/cobra"
)

var KeyStorageModeCmd = &cobra.Command{
	Use:   "storage-mode",
	Short: "See and manage SOPS key storage modes",
	Long:  `See and manage SOPS key storage modes`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.ExecuteCobraCommand(domain.KeyStorageMode, cmd, args)
	},
}

func init() {
	pkg.InitCobraCommand(domain.KeyStorageMode, KeyStorageModeCmd)
}
