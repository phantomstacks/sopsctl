package list

import (
	"fmt"
	"sopsctl/pkg/domain"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type KeyListCmd struct {
	options *KeyListCmdOptions
	skm     domain.SopsKeyManager
}

func (k KeyListCmd) InitCmd(cmd *cobra.Command) {

	cmd.Flags().Bool("show-sensitive", false, "Show private keys in the list output")
}

func NewKeyListCmd(skm domain.SopsKeyManager) *KeyListCmd {
	return &KeyListCmd{skm: skm}
}

func (k KeyListCmd) UseOptions(cmd *cobra.Command, _ []string) (domain.CommandExecutor, error) {
	showSensitive, err := cmd.Flags().GetBool("show-sensitive")
	if err != nil {
		return nil, err
	}
	k.options = NewKeyListCmdOptions(showSensitive)

	return k, nil
}

func (k KeyListCmd) Execute() (string, error) {
	keys, err := k.skm.ListContextsWithKeys()
	if err != nil {
		return "", fmt.Errorf("list keys: %w", err)
	}
	if len(keys) == 0 {
		return color.YellowString("No SOPS keys found."), nil
	}
	output := color.GreenString("SOPS Keys found for contexts:\n")
	for _, ctx := range keys {
		output += "- " + color.CyanString(ctx) + ": "
		publicKey, _ := k.skm.GetPublicKey(ctx)
		privateKey, _ := k.skm.GetPrivateKey(ctx)
		output += "\n  "
		if publicKey != "" {
			output += " Public Key: " + color.GreenString(publicKey)
		} else {
			output += " Public Key: " + color.RedString("<not set>")
		}

		if k.options.ShowSensitive {
			output += "\n  "
			if privateKey != "" {
				output += " Private Key: " + color.GreenString(privateKey)
			} else {
				output += " Private Key: " + color.RedString("<not set>")
			}
		}
		output += "\n"
	}
	return output, nil

}
