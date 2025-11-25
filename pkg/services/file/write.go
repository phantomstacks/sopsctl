package file

import (
	"fmt"
	"os"
	"path/filepath"
)

func AtomicWriteFile(dest string, data []byte) error {
	dir := filepath.Dir(dest)
	// Create temp file with a prefix to identify origin; pattern must include a '*' per CreateTemp docs.
	tmpFile, err := os.CreateTemp(dir, ".pfux-edit-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	// Ensure cleanup on any exit path.
	tmpName := tmpFile.Name()
	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}

	// Write contents
	if _, err = tmpFile.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}
	// Sync file content to storage
	if err = tmpFile.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync temp file: %w", err)
	}
	// Set desired permissions (CreateTemp may use 0600 masked by umask)
	const fileWritePermission = 0600
	if err = tmpFile.Chmod(fileWritePermission); err != nil {
		cleanup()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	// Close before rename (some platforms require this for durability guarantees)
	if err = tmpFile.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}

	// Rename is atomic on POSIX when source & dest are on same filesystem and in same directory.
	if err = os.Rename(tmpName, dest); err != nil {
		cleanup()
		return fmt.Errorf("rename temp file: %w", err)
	}

	// Optionally sync directory to ensure metadata (rename) is durable.
	if dirHandle, err2 := os.Open(dir); err2 == nil {
		_ = dirHandle.Sync() // ignore error; best-effort
		_ = dirHandle.Close()
	}
	return nil
}
