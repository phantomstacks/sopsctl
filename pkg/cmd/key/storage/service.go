package storage

import (
	"fmt"
	"phantom-flux/pkg/cmd/help"
	"phantom-flux/pkg/domain"

	"github.com/spf13/cobra"
)

const setStorageModeFlagName = "set-storage-mode"

type keyRootCmdOptions struct {
	StorageMode domain.StorageMode
}

type KeyStorageModeCmd struct {
	options *keyRootCmdOptions
	storage domain.KeyStorage
}

func NewKeyStorageModeCmd(storage domain.KeyStorage) *KeyStorageModeCmd {
	return &KeyStorageModeCmd{storage: storage}
}

func (k KeyStorageModeCmd) UseOptions(cmd *cobra.Command, args []string) (domain.CommandExecutor, error) {
	storageModeStr, err := cmd.Flags().GetString(setStorageModeFlagName)
	if err != nil {
		return nil, err
	}
	if storageModeStr == "" {
		return help.NewHelpExecutor(cmd), nil
	}

	sm := domain.StorageMode(storageModeStr)
	if !sm.IsValid() {
		return nil, fmt.Errorf("invalid storage mode: %s", storageModeStr)
	}

	options := &keyRootCmdOptions{
		StorageMode: sm,
	}
	k.options = options // to be used in Execute
	return k, nil
}

func (k KeyStorageModeCmd) InitCmd(cmd *cobra.Command) {
	cmd.Flags().StringP(setStorageModeFlagName, "s", "", "Storage mode for SOPS keys (local, cluster.)")
}

func (k KeyStorageModeCmd) Execute() (string, error) {
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
