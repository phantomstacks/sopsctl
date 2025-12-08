package add

import (
	"fmt"
	"sopsctl/pkg/domain"
	"sopsctl/pkg/services/utils"

	"github.com/spf13/cobra"
)

type KeyAddCmd struct {
	options          KeyAddCmdOptions
	secretKeyManager domain.SopsKeyManager
}

func (k KeyAddCmd) InitCmd(cmd *cobra.Command) {
	cmd.Flags().Bool("from-current-context", false, "Use the current Kubernetes context instead of specifying --context")
	cmd.Flags().StringP("namespace", "n", "flux-system", "The namespace where the secret is located")
	cmd.Flags().StringP("secret", "s", "sops-age", "The name of the secret containing the SOPS key")
	cmd.Flags().StringP("key", "k", "age.agekey", "The key within the secret that holds the SOPS key")
}

func NewKeyAddCmd(secretKeyManager domain.SopsKeyManager) *KeyAddCmd {
	return &KeyAddCmd{secretKeyManager: secretKeyManager}
}

func (k KeyAddCmd) UseOptions(cmd *cobra.Command, args []string) (domain.CommandExecutor, error) {
	// mark args as intentionally unused beyond validation
	_ = args
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected positional arguments: %v", args)
	}
	gFlags, err := utils.UseGlobalFlags(cmd)
	if err != nil {
		return nil, err
	}

	namespace, _ := cmd.Flags().GetString("namespace")
	secretName, _ := cmd.Flags().GetString("secret")
	secretKey, _ := cmd.Flags().GetString("key")

	k.options = *NewKeyAddCmdOptions(gFlags.Cluster, namespace, secretName, secretKey)
	return k, nil
}

func (k KeyAddCmd) Execute() (string, error) {
	result, err := k.secretKeyManager.AddKeyFromCluster(k.options.Cluster, k.options.Namespace, k.options.SecretName, k.options.SecretKey)
	if err != nil {
		return "", err
	}
	return result, nil
}
