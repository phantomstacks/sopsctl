package remove

import (
	"fmt"
	"phantom-flux/pkg/domain"
	"slices"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type KeyRemoveCmd struct {
	options *KeyRemoveCmdOptions
	skm     domain.SopsKeyManager
}

func (k KeyRemoveCmd) InitCmd(cmd *cobra.Command) {
	cmd.Flags().Bool("all", false, "Remove all SOPS keys from local storage")
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Use = "remove [cluster-name]"
}

func NewKeyRemoveCmd(skm domain.SopsKeyManager) *KeyRemoveCmd {
	return &KeyRemoveCmd{skm: skm}
}

func (k KeyRemoveCmd) UseOptions(cmd *cobra.Command, args []string) (domain.CommandExecutor, error) {
	removeAll, err := cmd.Flags().GetBool("all")
	if err != nil {
		return nil, err
	}
	k.options = NewKeyRemoveCmdOptions(removeAll, args)

	return k, nil
}

func (k KeyRemoveCmd) Execute() (string, error) {
	keys, err := k.skm.ListContextsWithKeys()
	if err != nil {
		return "", fmt.Errorf("list keys: %w", err)
	}
	if len(keys) == 0 {
		return color.YellowString("No SOPS keys found."), nil
	}
	var output string
	for _, ctx := range keys {
		isInArgs := slices.Contains(k.options.ClusterNames, ctx)
		if k.options.RemoveAll || isInArgs {
			err := k.skm.RemoveKeyForContext(ctx)
			if err != nil {
				return "", fmt.Errorf("failed to remove SOPS key for context %s: %w", ctx, err)
			}
			output += fmt.Sprintf("Removed SOPS key for context: %s\n", color.CyanString(ctx))
		}

	}
	return output, nil
}
