package decrypt

import (
	"sopsctl/pkg/domain"
	"sopsctl/pkg/services/utils"

	"github.com/spf13/cobra"
)

type SecretDecryptCmd struct {
	options           *SecretDecryptOptions
	keyManager        domain.SopsKeyManager
	encryptionService domain.EncryptionService
}

func (d SecretDecryptCmd) InitCmd(_ *cobra.Command) {
	//TODO implement me
	panic("implement me")
}

func NewSecretDecryptCmd(keyManager domain.SopsKeyManager, encryptionService domain.EncryptionService) *SecretDecryptCmd {
	return &SecretDecryptCmd{keyManager: keyManager, encryptionService: encryptionService}
}

func (d SecretDecryptCmd) UseOptions(cmd *cobra.Command, args []string) (domain.CommandExecutor, error) {
	gFlags, err := utils.UseGlobalFlags(cmd)
	if err != nil {
		return nil, err
	}
	check, err := utils.UserFileArg(args)
	if err != nil {
		return nil, err
	}
	d.options = NewSecretDecryptOptions(check, gFlags.Cluster)
	return d, nil
}

func (d SecretDecryptCmd) Execute() (string, error) {
	var output string
	privateKey, err := d.keyManager.GetPrivateKey(d.options.Cluster)
	if err != nil {
		return "", err
	}
	decrypted, err := d.encryptionService.Decrypt(d.options.FilePath, privateKey)
	if err != nil {
		return "", err
	}
	output = string(decrypted)
	return output, nil
}
