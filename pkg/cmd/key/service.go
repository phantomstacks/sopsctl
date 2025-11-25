package key

import (
	"fmt"
	"phantom-flux/pkg/domain"

	"github.com/spf13/cobra"
)

const setStorageModeFlagName = "set-storage-mode"

type keyRootCmdOptions struct {
	StorageMode domain.StorageMode
}

type KeyRootCmd struct {
	options *keyRootCmdOptions
	storage domain.KeyStorage
}

func (k KeyRootCmd) UseOptions(cmd *cobra.Command, args []string) (domain.CommandExecutor, error) {
	storageModeStr, err := cmd.Flags().GetString(setStorageModeFlagName)
	if err != nil {
		return nil, err
	}
	options := &keyRootCmdOptions{
		StorageMode: domain.StorageMode(storageModeStr),
	}
	k.options = options // to be used in Execute
	return k, nil
}

func (k KeyRootCmd) InitCmd(cmd *cobra.Command) {
	cmd.Flags().StringP(setStorageModeFlagName, "s", "file", "Storage mode for SOPS keys (local, in-cluster.)")
}

func (k KeyRootCmd) Execute() (string, error) {
	currentMode, err := k.storage.GetStorageMode()
	if err != nil {
		return "", err
	}
	if currentMode.ToString() == k.options.StorageMode.ToString() {
		return "", nil
	}
	err = k.storage.SetStorageMode(k.options.StorageMode)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("key storage mode set to %s", k.options.StorageMode.ToString()), nil
}
