// Package cmd /*
package cmd

import (
	"phantom-flux/cmd/key_commands"

	"github.com/spf13/cobra"
)

// sopsCmd represents the sops command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(keyCmd)
	keyCmd.AddCommand(key_commands.KeyAddCmd)
	keyCmd.AddCommand(key_commands.KeyListCmd)
	keyCmd.AddCommand(key_commands.RemoveCmd)
	keyCmd.AddCommand(key_commands.KeyStorageModeCmd)
}
