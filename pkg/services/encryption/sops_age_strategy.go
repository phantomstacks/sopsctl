package encryption

import (
	"fmt"
	"sopsctl/pkg/domain"

	"filippo.io/age"
	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/aes"
	keysource "github.com/getsops/sops/v3/age"
	"github.com/getsops/sops/v3/cmd/sops/common"
	"github.com/getsops/sops/v3/cmd/sops/formats"
	"github.com/getsops/sops/v3/config"
	"github.com/getsops/sops/v3/decrypt"
	"github.com/getsops/sops/v3/keys"
	"github.com/getsops/sops/v3/keyservice"

	"os"
	"time"
)

type SopsAgeDecryptStrategy struct {
	checkSopsMac bool
}

func (s *SopsAgeDecryptStrategy) EncryptData(data []byte, publicKey string) ([]byte, error) {
	store := common.StoreForFormat(formats.Yaml, config.NewStoresConfig())
	branches, err := store.LoadPlainFile(data)
	if err != nil {
		panic(err)
	}
	masterKey, err := keysource.MasterKeyFromRecipient(publicKey)
	if err != nil {
		panic(err)
	}
	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			KeyGroups: []sops.KeyGroup{
				[]keys.MasterKey{masterKey},
			},
			EncryptedRegex: "^(data|stringData)$",
		},
	}

	dataKey, errs := tree.GenerateDataKeyWithKeyServices(
		[]keyservice.KeyServiceClient{keyservice.NewLocalClient()},
	)
	if errs != nil {
		return nil, fmt.Errorf("failed to generate data key: %v", errs)
	}
	err = common.EncryptTree(common.EncryptTreeOpts{
		DataKey: dataKey,
		Tree:    &tree,
		Cipher:  aes.NewCipher(),
	})
	if err != nil {
		return nil, err
	}
	result, err := store.EmitEncryptedFile(tree)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewSopsAgeDecryptStrategy() domain.EncryptionService {
	return &SopsAgeDecryptStrategy{
		checkSopsMac: false,
	}
}

func (s *SopsAgeDecryptStrategy) Decrypt(filePath, ageKey string) ([]byte, error) {
	// parse the private Age key (optional but good for validation)

	_, err := age.ParseX25519Identity(ageKey)
	if err != nil {
		return nil, fmt.Errorf("bad age key: %w", err)
	}
	err = os.Setenv("SOPS_AGE_KEY", ageKey)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.Unsetenv("SOPS_AGE_KEY")
		if err != nil {
			panic(err)
		}
	}()
	data, err := os.ReadFile(filePath)

	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return s.SopsDecryptWithFormat(data, formats.Yaml, formats.Yaml)
}

func (s *SopsAgeDecryptStrategy) DecryptData(data []byte, ageKey string) ([]byte, error) {
	// parse the private Age key (optional but good for validation)
	_, err := age.ParseX25519Identity(ageKey)
	if err != nil {
		return nil, fmt.Errorf("bad age key: %w", err)
	}

	err = os.Setenv("SOPS_AGE_KEY", ageKey)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Unsetenv("SOPS_AGE_KEY")
	}()

	return decrypt.Data(data, "yaml")
}

func (s *SopsAgeDecryptStrategy) SopsDecryptWithFormat(data []byte, inputFormat, outputFormat formats.Format) (_ []byte, err error) {
	store := common.StoreForFormat(inputFormat, config.NewStoresConfig())

	tree, err := store.LoadEncryptedFile(data)
	if err != nil {
		return nil, err
	}

	metadataKey, err := tree.Metadata.GetDataKey()
	if err != nil {
		return nil, err
	}

	cipher := aes.NewCipher()
	mac, err := tree.Decrypt(metadataKey, cipher)
	if err != nil {
		return nil, err
	}

	if s.checkSopsMac {
		// Compute the hash of the cleartext tree and compare it with
		// the one that was stored in the document. If they match,
		// integrity was preserved
		// Ref: github.com/getsops/sops/v3/decrypt/decrypt.go
		originalMac, err := cipher.Decrypt(
			tree.Metadata.MessageAuthenticationCode,
			metadataKey,
			tree.Metadata.LastModified.Format(time.RFC3339),
		)
		if err != nil {
			return nil, err
		}
		if originalMac != mac {
			// If the file has an empty MAC, display "no MAC"
			if originalMac == "" {
				originalMac = "no MAC"
			}
			return nil, fmt.Errorf("failed to verify sops data integrity: expected mac '%s', got '%s'", originalMac, mac)
		}
	}

	outputStore := common.StoreForFormat(outputFormat, config.NewStoresConfig())
	out, err := outputStore.EmitPlainFile(tree.Branches)
	if err != nil {
		return nil, err
	}
	return out, err
}

func (s *SopsAgeDecryptStrategy) EncryptFile(filePath string, publicKey string) ([]byte, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return s.EncryptData(file, publicKey)
}
