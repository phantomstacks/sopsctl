package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sopsctl/pkg/services/helpers"
)

type WriterService struct {
}

func NewFileService() *WriterService {
	return &WriterService{}
}

func (s *WriterService) CreateTempFile(decrypted []byte) (string, func(), error) {
	tempDir := filepath.Join(os.TempDir(), "sopsctl-edit-"+helpers.RandomString(10))
	cleanup := func() {
		//if not excised, ignore errors
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			if os.IsNotExist(cleanupErr) {
				return
			}
			helpers.PrintError("failed to clean up temp directory: %v", cleanupErr)
		}
	}
	const tempDirPermission = 0700
	if err := os.MkdirAll(tempDir, tempDirPermission); err != nil {
		return "", cleanup, fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempFilePath := filepath.Join(tempDir, "decrypted_secret.yaml")
	if err := AtomicWriteFile(tempFilePath, decrypted); err != nil {
		return "", cleanup, fmt.Errorf("failed to write to temp file: %w", err)
	}
	return tempFilePath, cleanup, nil
}
