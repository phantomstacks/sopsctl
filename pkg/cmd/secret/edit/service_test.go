package edit

import (
	"errors"
	"io"
	"testing"

	"filippo.io/age"
	"github.com/getsops/sops/v3/cmd/sops/formats"
)

// Mock implementations for testing

type mockKeyManager struct {
	privateKey    string
	publicKey     string
	privateKeyErr error
	publicKeyErr  error
}

func (m *mockKeyManager) GetPrivateKey(_ string) (string, error) {
	return m.privateKey, m.privateKeyErr
}

func (m *mockKeyManager) GetPublicKey(_ string) (string, error) {
	return m.publicKey, m.publicKeyErr
}

func (m *mockKeyManager) GetIdentityCurrentCtx() (age.Identity, error) {
	return nil, nil
}

func (m *mockKeyManager) AddKeyFromCluster(_, _, _, _ string) (string, error) {
	return "", nil
}

func (m *mockKeyManager) ListContextsWithKeys() ([]string, error) {
	return nil, nil
}

func (m *mockKeyManager) RemoveKeyForContext(_ string) error {
	return nil
}

type mockEncryptionService struct {
	decryptedData []byte
	encryptedData []byte
	decryptErr    error
	encryptErr    error
}

func (m *mockEncryptionService) Decrypt(_, _ string) ([]byte, error) {
	return m.decryptedData, m.decryptErr
}

func (m *mockEncryptionService) DecryptData(_ []byte, _ string) ([]byte, error) {
	return m.decryptedData, m.decryptErr
}

func (m *mockEncryptionService) SopsDecryptWithFormat(_ []byte, _, _ formats.Format) ([]byte, error) {
	return nil, nil
}

func (m *mockEncryptionService) EncryptFile(_, _ string) ([]byte, error) {
	return m.encryptedData, m.encryptErr
}

func (m *mockEncryptionService) EncryptData(_ []byte, _ string) ([]byte, error) {
	return m.encryptedData, m.encryptErr
}

type mockDecoder struct {
	defaultKey     string
	decodedData    []byte
	reEncodeFunc   func([]byte) ([]byte, error)
	defaultKeyErr  error
	editDecodedErr error
}

func (m *mockDecoder) GetDefaultKey(_ []byte) (string, error) {
	return m.defaultKey, m.defaultKeyErr
}

func (m *mockDecoder) EditDecodedFile(_ []byte, _ string) ([]byte, func([]byte) ([]byte, error), error) {
	if m.reEncodeFunc == nil {
		m.reEncodeFunc = func(b []byte) ([]byte, error) {
			return b, nil
		}
	}
	return m.decodedData, m.reEncodeFunc, m.editDecodedErr
}

func (m *mockDecoder) CountDecodedFileEntries(_ []byte) (int, error) {
	return 0, nil
}

type mockEditor struct {
	editedContent []byte
	editErr       error
}

func (m *mockEditor) EditFile(_ string) ([]byte, error) {
	return m.editedContent, m.editErr
}

func (m *mockEditor) EditTempFile(_, _ string, _ []byte) ([]byte, func(), error) {
	return nil, func() {}, nil
}

func (m *mockEditor) EditStream(_, _ string, _ io.Reader) ([]byte, error) {
	return nil, nil
}

func (m *mockEditor) EditFileWithPostEditCallback(_ string, _ func([]byte) ([]byte, error)) ([]byte, error) {
	return nil, nil
}

type mockFileService struct {
	tempFilePath string
	cleanupFunc  func()
	createErr    error
}

func (m *mockFileService) CreateTempFile(_ []byte) (string, func(), error) {
	if m.cleanupFunc == nil {
		m.cleanupFunc = func() {}
	}
	return m.tempFilePath, m.cleanupFunc, m.createErr
}

// Test decryptFile method

func TestDecryptFile_Success(t *testing.T) {
	expectedData := []byte("decrypted content")
	mockKM := &mockKeyManager{
		privateKey: "test-private-key",
	}
	mockEnc := &mockEncryptionService{
		decryptedData: expectedData,
	}

	cmd := SecretEditCmd{
		keyManager:        mockKM,
		encryptionService: mockEnc,
		options: &editCmdOptions{
			File:    "test.yaml",
			Cluster: "test-cluster",
		},
	}

	result, err := cmd.decryptFile()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(result) != string(expectedData) {
		t.Errorf("Expected %s, got %s", expectedData, result)
	}
}

func TestDecryptFile_PrivateKeyError(t *testing.T) {
	expectedErr := errors.New("key not found")
	mockKM := &mockKeyManager{
		privateKeyErr: expectedErr,
	}

	cmd := SecretEditCmd{
		keyManager: mockKM,
		options: &editCmdOptions{
			Cluster: "test-cluster",
		},
	}

	_, err := cmd.decryptFile()

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestDecryptFile_DecryptionError(t *testing.T) {
	expectedErr := errors.New("decryption failed")
	mockKM := &mockKeyManager{
		privateKey: "test-key",
	}
	mockEnc := &mockEncryptionService{
		decryptErr: expectedErr,
	}

	cmd := SecretEditCmd{
		keyManager:        mockKM,
		encryptionService: mockEnc,
		options: &editCmdOptions{
			File:    "test.yaml",
			Cluster: "test-cluster",
		},
	}

	_, err := cmd.decryptFile()

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}

// Test decodeIfNeeded method

func TestDecodeIfNeeded_NoDecodeRequired(t *testing.T) {
	inputData := []byte("original data")
	cmd := SecretEditCmd{
		options: &editCmdOptions{
			ShouldDecodeAsFile: false,
		},
	}

	result, reEncodeFunc, err := cmd.decodeIfNeeded(inputData)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(result) != string(inputData) {
		t.Errorf("Expected %s, got %s", inputData, result)
	}

	// Test that reEncodeFunc is a no-op
	encoded, err := reEncodeFunc([]byte("test"))
	if err != nil {
		t.Errorf("Expected no error from reEncodeFunc, got %v", err)
	}
	if string(encoded) != "test" {
		t.Error("Expected reEncodeFunc to be a no-op")
	}
}

func TestDecodeIfNeeded_WithDecodeExplicitKey(t *testing.T) {
	inputData := []byte("encrypted data")
	decodedData := []byte("decoded data")
	mockDec := &mockDecoder{
		decodedData: decodedData,
	}

	cmd := SecretEditCmd{
		decoder: mockDec,
		options: &editCmdOptions{
			ShouldDecodeAsFile: true,
			DecodeAsFileKey:    "myKey",
		},
	}

	result, _, err := cmd.decodeIfNeeded(inputData)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(result) != string(decodedData) {
		t.Errorf("Expected %s, got %s", decodedData, result)
	}
}

func TestDecodeIfNeeded_WithDecodeDefaultKey(t *testing.T) {
	inputData := []byte("encrypted data")
	decodedData := []byte("decoded data")
	mockDec := &mockDecoder{
		defaultKey:  "autoKey",
		decodedData: decodedData,
	}

	cmd := SecretEditCmd{
		decoder: mockDec,
		options: &editCmdOptions{
			ShouldDecodeAsFile: true,
			DecodeAsFileKey:    "", // Empty means use default
		},
	}

	result, _, err := cmd.decodeIfNeeded(inputData)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(result) != string(decodedData) {
		t.Errorf("Expected %s, got %s", decodedData, result)
	}
}

func TestDecodeIfNeeded_DecodeError(t *testing.T) {
	expectedErr := errors.New("decode failed")
	mockDec := &mockDecoder{
		defaultKey:     "key",
		editDecodedErr: expectedErr,
	}

	cmd := SecretEditCmd{
		decoder: mockDec,
		options: &editCmdOptions{
			ShouldDecodeAsFile: true,
		},
	}

	_, _, err := cmd.decodeIfNeeded([]byte("data"))

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}

// Test resolveDecodeKey method

func TestResolveDecodeKey_ExplicitKey(t *testing.T) {
	cmd := SecretEditCmd{
		options: &editCmdOptions{
			DecodeAsFileKey: "myExplicitKey",
		},
	}

	result, err := cmd.resolveDecodeKey([]byte("data"))

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "myExplicitKey" {
		t.Errorf("Expected 'myExplicitKey', got '%s'", result)
	}
}

func TestResolveDecodeKey_DefaultKey(t *testing.T) {
	mockDec := &mockDecoder{
		defaultKey: "defaultKey",
	}

	cmd := SecretEditCmd{
		decoder: mockDec,
		options: &editCmdOptions{
			DecodeAsFileKey: "",
		},
	}

	result, err := cmd.resolveDecodeKey([]byte("data"))

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "defaultKey" {
		t.Errorf("Expected 'defaultKey', got '%s'", result)
	}
}

func TestResolveDecodeKey_DefaultKeyError(t *testing.T) {
	expectedErr := errors.New("no default key found")
	mockDec := &mockDecoder{
		defaultKeyErr: expectedErr,
	}

	cmd := SecretEditCmd{
		decoder: mockDec,
		options: &editCmdOptions{
			DecodeAsFileKey: "",
		},
	}

	_, err := cmd.resolveDecodeKey([]byte("data"))

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}

// Test editInTempFile method

func TestEditInTempFile_Success(t *testing.T) {
	inputContent := []byte("original content")
	editedContent := []byte("edited content")

	mockFS := &mockFileService{
		tempFilePath: "/tmp/test.yaml",
	}
	mockEd := &mockEditor{
		editedContent: editedContent,
	}

	cmd := SecretEditCmd{
		fileService: mockFS,
		editor:      mockEd,
	}

	result, err := cmd.editInTempFile(inputContent)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(result) != string(editedContent) {
		t.Errorf("Expected %s, got %s", editedContent, result)
	}
}

func TestEditInTempFile_CreateTempFileError(t *testing.T) {
	expectedErr := errors.New("failed to create temp file")
	mockFS := &mockFileService{
		createErr: expectedErr,
	}

	cmd := SecretEditCmd{
		fileService: mockFS,
	}

	_, err := cmd.editInTempFile([]byte("data"))

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestEditInTempFile_EditorError(t *testing.T) {
	expectedErr := errors.New("editor failed")
	mockFS := &mockFileService{
		tempFilePath: "/tmp/test.yaml",
	}
	mockEd := &mockEditor{
		editErr: expectedErr,
	}

	cmd := SecretEditCmd{
		fileService: mockFS,
		editor:      mockEd,
	}

	_, err := cmd.editInTempFile([]byte("data"))

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}

// Test encryptAndSave method

func TestEncryptAndSave_Success(t *testing.T) {
	editedContent := []byte("edited content")
	encryptedData := []byte("encrypted data")

	mockKM := &mockKeyManager{
		publicKey: "test-public-key",
	}
	mockEnc := &mockEncryptionService{
		encryptedData: encryptedData,
	}

	// Track if AtomicWriteFile was called
	var writeCalled bool
	originalWrite := atomicWriteFile
	atomicWriteFile = func(path string, data []byte) error {
		writeCalled = true
		if string(data) != string(encryptedData) {
			t.Errorf("Expected to write %s, got %s", encryptedData, data)
		}
		return nil
	}
	defer func() { atomicWriteFile = originalWrite }()

	cmd := SecretEditCmd{
		keyManager:        mockKM,
		encryptionService: mockEnc,
		options: &editCmdOptions{
			File:    "test.yaml",
			Cluster: "test-cluster",
		},
	}

	reEncodeFunc := func(b []byte) ([]byte, error) {
		return b, nil
	}

	err := cmd.encryptAndSave(editedContent, reEncodeFunc)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !writeCalled {
		t.Error("Expected AtomicWriteFile to be called")
	}
}

func TestEncryptAndSave_PublicKeyError(t *testing.T) {
	expectedErr := errors.New("public key not found")
	mockKM := &mockKeyManager{
		publicKeyErr: expectedErr,
	}

	cmd := SecretEditCmd{
		keyManager: mockKM,
		options: &editCmdOptions{
			Cluster: "test-cluster",
		},
	}

	reEncodeFunc := func(b []byte) ([]byte, error) {
		return b, nil
	}

	err := cmd.encryptAndSave([]byte("data"), reEncodeFunc)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestEncryptAndSave_ReEncodeError(t *testing.T) {
	expectedErr := errors.New("re-encode failed")
	mockKM := &mockKeyManager{
		publicKey: "test-key",
	}

	cmd := SecretEditCmd{
		keyManager: mockKM,
		options: &editCmdOptions{
			Cluster: "test-cluster",
		},
	}

	reEncodeFunc := func(b []byte) ([]byte, error) {
		return nil, expectedErr
	}

	err := cmd.encryptAndSave([]byte("data"), reEncodeFunc)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestEncryptAndSave_EncryptionError(t *testing.T) {
	expectedErr := errors.New("encryption failed")
	mockKM := &mockKeyManager{
		publicKey: "test-key",
	}
	mockEnc := &mockEncryptionService{
		encryptErr: expectedErr,
	}

	cmd := SecretEditCmd{
		keyManager:        mockKM,
		encryptionService: mockEnc,
		options: &editCmdOptions{
			Cluster: "test-cluster",
		},
	}

	reEncodeFunc := func(b []byte) ([]byte, error) {
		return b, nil
	}

	err := cmd.encryptAndSave([]byte("data"), reEncodeFunc)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}
