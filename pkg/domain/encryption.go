package domain

import "github.com/getsops/sops/v3/cmd/sops/formats"

type EncryptionService interface {
	Decrypt(filePath, ageKey string) ([]byte, error)
	DecryptData(data []byte, ageKey string) ([]byte, error)
	SopsDecryptWithFormat(data []byte, inputFormat, outputFormat formats.Format) (_ []byte, err error)
	EncryptFile(filePath string, publicKey string) ([]byte, error)
	EncryptData(data []byte, publicKey string) ([]byte, error)
}
