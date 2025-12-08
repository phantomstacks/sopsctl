package edit

import (
	"fmt"
	"sopsctl/pkg/domain"
	"sopsctl/pkg/services/file"
	"sopsctl/pkg/services/utils"

	"github.com/spf13/cobra"
)

const (
	decodeFlagName      = "decode"
	decodeKey           = "k"
	decodeAsEnvFlagName = "env"
)

type SecretEditCmd struct {
	options           *editCmdOptions
	keyManager        domain.SopsKeyManager
	encryptionService domain.EncryptionService
	decoder           domain.Base64Decoder
	editor            domain.UserEditorService
	fileService       domain.FileService
}

func (e SecretEditCmd) InitCmd(cmd *cobra.Command) {
	cmd.Flags().BoolP(decodeFlagName, "d", false,
		`Allows you to edit a decoded secret property in place without needing to manually decrypt and re-encrypt the entire file.
If used without any additional arguments, there must be exactly one secret property in the file.

You can also specify the key within the secret to edit -d -k <key>.
`)
	cmd.Flags().StringP(decodeKey, "k", "", "Specifies the key within the secret to decode and edit.")
	cmd.Flags().BoolP(decodeAsEnvFlagName, "e", false, "Specifies the environment variable that holds the decoded value.")
}

// NewSecretEditCmd Updated constructor with dependencies for DI container.
func NewSecretEditCmd(keyManager domain.SopsKeyManager, encryptionService domain.EncryptionService, decoder domain.Base64Decoder, editor domain.UserEditorService, fileService domain.FileService) domain.CommandBuilder {
	return &SecretEditCmd{
		keyManager:        keyManager,
		encryptionService: encryptionService,
		decoder:           decoder,
		editor:            editor,
		fileService:       fileService,
	}
}

// Execute orchestrates the edit workflow: decrypt, decode (if needed), edit, encode, encrypt, and save.
func (e SecretEditCmd) Execute() (string, error) {
	decrypted, err := e.decryptFile()
	if err != nil {
		return "", err
	}

	decrypted, reEncodeFunc, err := e.decodeIfNeeded(decrypted)
	if err != nil {
		return "", err
	}
	copyOfDecrypted := &decrypted
	editedContent, err := e.editInTempFile(decrypted)
	if err != nil {
		return "", err
	}
	// If no changes were made, exit early
	if string(editedContent) == string(*copyOfDecrypted) {
		return "No changes made to the file", nil
	}

	if err := e.encryptAndSave(editedContent, reEncodeFunc); err != nil {
		return "", err
	}

	return "File edited and encrypted successfully", nil
}

// decryptFile retrieves the private key and decrypts the target file.
func (e SecretEditCmd) decryptFile() ([]byte, error) {
	privateKey, err := e.keyManager.GetPrivateKey(e.options.Cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key for cluster %s: %w", e.options.Cluster, err)
	}

	decrypted, err := e.encryptionService.Decrypt(e.options.File, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file %s: %w", e.options.File, err)
	}

	return decrypted, nil
}

// decodeIfNeeded optionally decodes a base64-encoded value within the decrypted data.
// Returns the data to edit and a re-encoding function to apply after editing.
func (e SecretEditCmd) decodeIfNeeded(decrypted []byte) ([]byte, func([]byte) ([]byte, error), error) {
	// Default re-encode function (no-op)
	reEncodeFunc := func(bytes []byte) ([]byte, error) {
		return bytes, nil
	}

	if !e.options.ShouldDecodeAsFile {
		return decrypted, reEncodeFunc, nil
	}

	valueKey, err := e.resolveDecodeKey(decrypted)
	if err != nil {
		return nil, nil, err
	}

	decodedData, reEncodeFunc, err := e.decoder.EditDecodedFile(decrypted, valueKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode file: %w", err)
	}

	return decodedData, reEncodeFunc, nil
}

// resolveDecodeKey determines which key to decode, using the default if not specified.
func (e SecretEditCmd) resolveDecodeKey(decrypted []byte) (string, error) {
	if e.options.DecodeAsFileKey != "" {
		return e.options.DecodeAsFileKey, nil
	}

	valueKey, err := e.decoder.GetDefaultKey(decrypted)
	if err != nil {
		return "", fmt.Errorf("failed to get default key: %w", err)
	}

	return valueKey, nil
}

// editInTempFile creates a temporary file with the content, opens it in an editor, and returns the edited content.
func (e SecretEditCmd) editInTempFile(content []byte) ([]byte, error) {
	tempFilePath, cleanup, err := e.fileService.CreateTempFile(content)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer cleanup()

	editedContent, err := e.editor.EditFile(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to edit file %s: %w", tempFilePath, err)
	}

	return editedContent, nil
}

// encryptAndSave re-encodes (if needed), encrypts the content, and writes it back to the original file.
func (e SecretEditCmd) encryptAndSave(editedContent []byte, reEncodeFunc func([]byte) ([]byte, error)) error {
	publicKey, err := e.keyManager.GetPublicKey(e.options.Cluster)
	if err != nil {
		return fmt.Errorf("failed to get SOPS public key for cluster %s: %w", e.options.Cluster, err)
	}

	encodedData, err := reEncodeFunc(editedContent)
	if err != nil {
		return fmt.Errorf("failed to re-encode data: %w", err)
	}

	encrypted, err := e.encryptionService.EncryptData(encodedData, publicKey)
	if err != nil {
		return fmt.Errorf("failed to re-encrypt file: %w", err)
	}

	if err := atomicWriteFile(e.options.File, encrypted); err != nil {
		return fmt.Errorf("failed to atomically write encrypted data to %s: %w", e.options.File, err)
	}

	return nil
}

// atomicWriteFile is a variable to allow mocking in tests
var atomicWriteFile = file.AtomicWriteFile

func (e SecretEditCmd) UseOptions(cmd *cobra.Command, args []string) (domain.CommandExecutor, error) {
	global, err := utils.UseGlobalFlags(cmd)
	if err != nil {
		return nil, err
	}

	filePath, err := utils.UserFileArg(args)
	if err != nil {
		return nil, err
	}

	shouldDecodeAsFile, err := cmd.Flags().GetBool(decodeFlagName)
	if err != nil {
		return nil, err
	}

	shouldDecodeDataKey, err := cmd.Flags().GetString(decodeKey)
	if err != nil {
		return nil, err
	}

	shouldDecodeAsEnv, err := cmd.Flags().GetBool(decodeAsEnvFlagName)
	if err != nil {
		return nil, err
	}

	e.options = newEditCmdOptions(filePath, global.Cluster, shouldDecodeAsEnv, shouldDecodeAsFile, shouldDecodeDataKey)
	return e, nil
}
