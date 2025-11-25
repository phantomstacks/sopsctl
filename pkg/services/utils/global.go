package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"phantom-flux/pkg/services/helpers"

	"github.com/spf13/cobra"
)

type GlobalFlags struct {
	Cluster string
}

func resolveCluster(cluster *string) error {
	if *cluster == "" {
		*cluster, _ = helpers.GetCtxNameFromCurrent()
		if *cluster == "" {
			err := fmt.Errorf("failed to get current context")
			helpers.PrintError("Failed to get current context: %v", err)
			return err
		}
	}
	return nil
}
func UseGlobalFlags(cmd *cobra.Command) (*GlobalFlags, error) {
	cluster := cmd.Flags().Lookup("cluster").Value.String()
	err := resolveCluster(&cluster)
	if err != nil {
		return nil, err
	}
	return &GlobalFlags{
		Cluster: cluster,
	}, nil
}

func UserFileArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("no file specified")
	}
	file := args[0]
	if file == "" {
		return "", fmt.Errorf("file argument is empty")
	}
	absoluteFilePath, err := filepath.Abs(file)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for file %s: %v", file, err)
	}
	// Check if file exists
	exists, err := os.Stat(absoluteFilePath)
	if os.IsNotExist(err) || exists.IsDir() {
		return "", fmt.Errorf("file does not exist: %s", absoluteFilePath)
	}
	return absoluteFilePath, nil

}
