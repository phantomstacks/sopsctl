package editor

import (
	"bytes"
	"os"
	"path/filepath"
	"phantom-flux/pkg/services/encryption"
	"strings"
	"testing"
)

const (
	testAgeKey       = "AGE-SECRET-KEY-13ZLWP4WFHQ6VHC2J5YYEUCFKGLZTD3SXQQPEGK3WU2M8FKYC238S7ZKNSV"
	testAgePublicKey = "age1qnswq576pku84s2wyw4kr59ywvvdzua6crtdz0sf0l9udnje6c5snqfc2d"
)

// TestEditEncryptedSecret tests the full workflow of decrypting, editing, and re-encrypting a secret
func TestEditEncryptedSecret(t *testing.T) {
	// Create encryption service
	encryptionSvc := encryption.NewSopsAgeDecryptStrategy()

	// Read the encrypted test file
	encryptedPath := filepath.Join("testdata", "enc.yaml")
	encryptedData, err := os.ReadFile(encryptedPath)
	if err != nil {
		t.Fatalf("Failed to read encrypted test file: %v", err)
	}

	// Decrypt the data
	decryptedData, err := encryptionSvc.DecryptData(encryptedData, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	// Verify decryption worked by checking for expected content
	if !strings.Contains(string(decryptedData), "SecretThings") {
		t.Errorf("Decrypted data doesn't contain expected secret value")
	}

	// Simulate editing by creating a mock editor that modifies the content
	editor := NewEditor("sed", "-i", "s/SecretThings/NewSecretValue/g")

	// Edit the decrypted content
	modifiedData, cleanup, err := editor.EditTempFile("secret-", ".yaml", decryptedData)
	if err != nil {
		t.Fatalf("Failed to edit temp file: %v", err)
	}
	defer cleanup()

	// Verify the edit was made
	if !strings.Contains(string(modifiedData), "NewSecretValue") {
		t.Errorf("Modified data doesn't contain edited value: %s", string(modifiedData))
	}

	// Re-encrypt the modified data
	reencryptedData, err := encryptionSvc.EncryptData(modifiedData, testAgePublicKey)
	if err != nil {
		t.Fatalf("Failed to re-encrypt data: %v", err)
	}

	// Verify re-encryption worked by checking for SOPS metadata
	reencryptedStr := string(reencryptedData)
	if !strings.Contains(reencryptedStr, "ENC[") {
		t.Errorf("Re-encrypted data doesn't contain encrypted markers")
	}
	if !strings.Contains(reencryptedStr, "sops:") {
		t.Errorf("Re-encrypted data doesn't contain SOPS metadata")
	}

	// Decrypt again to verify the new secret value
	finalDecrypted, err := encryptionSvc.DecryptData(reencryptedData, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt re-encrypted data: %v", err)
	}

	// Verify the edited value is present in the final decrypted data
	if !strings.Contains(string(finalDecrypted), "NewSecretValue") {
		t.Errorf("Final decrypted data doesn't contain the edited value")
	}
	if strings.Contains(string(finalDecrypted), "SecretThings") {
		t.Errorf("Final decrypted data still contains old value")
	}
}

// TestEditFileWithPostEditCallback tests editing with automatic encryption callback
func TestEditFileWithPostEditCallback(t *testing.T) {
	encryptionSvc := encryption.NewSopsAgeDecryptStrategy()

	// Read and decrypt the test file
	encryptedPath := filepath.Join("testdata", "enc.yaml")
	decryptedData, err := encryptionSvc.Decrypt(encryptedPath, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt test file: %v", err)
	}

	// Write decrypted data to a temp file for editing
	tmpFile, err := os.CreateTemp("", "test-edit-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(decryptedData); err != nil {
		tmpFile.Close()
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Create an editor that uses cat (no-op for testing)
	editor := NewEditor("cat")

	// Define a post-edit callback that encrypts the content
	postEditCallback := func(editedContent []byte) ([]byte, error) {
		return encryptionSvc.EncryptData(editedContent, testAgePublicKey)
	}

	// Edit file with callback
	encryptedResult, err := editor.EditFileWithPostEditCallback(tmpPath, postEditCallback)
	if err != nil {
		t.Fatalf("EditFileWithPostEditCallback failed: %v", err)
	}

	// Verify the result is encrypted
	if !strings.Contains(string(encryptedResult), "ENC[") {
		t.Errorf("Post-edit callback didn't encrypt the content")
	}
	if !strings.Contains(string(encryptedResult), "sops:") {
		t.Errorf("Post-edit callback result doesn't contain SOPS metadata")
	}
}

// TestDecryptEditEncryptWorkflow tests a complete workflow simulating user secret editing
func TestDecryptEditEncryptWorkflow(t *testing.T) {
	encryptionSvc := encryption.NewSopsAgeDecryptStrategy()

	// Step 1: Read encrypted secret
	encryptedPath := filepath.Join("testdata", "enc.yaml")
	encryptedContent, err := os.ReadFile(encryptedPath)
	if err != nil {
		t.Fatalf("Failed to read encrypted file: %v", err)
	}

	// Step 2: Decrypt
	decryptedContent, err := encryptionSvc.DecryptData(encryptedContent, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Verify original content
	if !strings.Contains(string(decryptedContent), "big-3-frontend-development") {
		t.Errorf("Decrypted content missing expected metadata name")
	}
	if !strings.Contains(string(decryptedContent), "SecretThings") {
		t.Errorf("Decrypted content missing expected secret data")
	}

	// Step 3: Edit content (simulate user changes)
	modifiedContent := bytes.Replace(
		decryptedContent,
		[]byte("SecretThings"),
		[]byte("UpdatedSecretValue"),
		1,
	)

	// Step 4: Re-encrypt
	reencryptedContent, err := encryptionSvc.EncryptData(modifiedContent, testAgePublicKey)
	if err != nil {
		t.Fatalf("Failed to re-encrypt: %v", err)
	}

	// Step 5: Verify encryption markers are present
	encStr := string(reencryptedContent)
	if !strings.Contains(encStr, "ENC[AES256_GCM,") {
		t.Errorf("Re-encrypted content missing encryption markers")
	}
	if !strings.Contains(encStr, "age:") {
		t.Errorf("Re-encrypted content missing age key metadata")
	}
	if !strings.Contains(encStr, testAgePublicKey) {
		t.Errorf("Re-encrypted content missing correct age recipient")
	}

	// Step 6: Decrypt again and verify changes
	finalDecrypted, err := encryptionSvc.DecryptData(reencryptedContent, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt final content: %v", err)
	}

	finalStr := string(finalDecrypted)
	if !strings.Contains(finalStr, "UpdatedSecretValue") {
		t.Errorf("Final decrypted content missing updated value")
	}
	if strings.Contains(finalStr, "SecretThings") {
		t.Errorf("Final decrypted content still contains old value")
	}

	// Verify structure is maintained
	if !strings.Contains(finalStr, "big-3-frontend-development") {
		t.Errorf("Final decrypted content lost metadata name")
	}
	if !strings.Contains(finalStr, "apiVersion: v1") {
		t.Errorf("Final decrypted content lost apiVersion")
	}
}

// TestEditStreamWithEncryption tests editing from a stream with encryption
func TestEditStreamWithEncryption(t *testing.T) {
	encryptionSvc := encryption.NewSopsAgeDecryptStrategy()

	// Read encrypted data
	encryptedPath := filepath.Join("testdata", "enc.yaml")
	encryptedData, err := os.ReadFile(encryptedPath)
	if err != nil {
		t.Fatalf("Failed to read encrypted file: %v", err)
	}

	// Decrypt
	decryptedData, err := encryptionSvc.DecryptData(encryptedData, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Create a reader from decrypted data
	reader := bytes.NewReader(decryptedData)

	// Use cat as no-op editor
	editor := NewEditor("cat")

	// Edit from stream
	editedData, err := editor.EditStream("stream-secret-", ".yaml", reader)
	if err != nil {
		t.Fatalf("EditStream failed: %v", err)
	}

	// Content should be unchanged (cat is no-op)
	if !bytes.Equal(decryptedData, editedData) {
		t.Errorf("EditStream modified content unexpectedly")
	}

	// Re-encrypt the edited data
	reencrypted, err := encryptionSvc.EncryptData(editedData, testAgePublicKey)
	if err != nil {
		t.Fatalf("Failed to re-encrypt: %v", err)
	}

	// Verify encryption
	if !strings.Contains(string(reencrypted), "sops:") {
		t.Errorf("Re-encrypted stream data missing SOPS metadata")
	}
}

// TestCompareDecryptedFiles verifies that our test files match
func TestCompareDecryptedFiles(t *testing.T) {
	encryptionSvc := encryption.NewSopsAgeDecryptStrategy()

	// Read the manually decrypted file
	decPath := filepath.Join("testdata", "dec.yaml")
	expectedDecrypted, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("Failed to read dec.yaml: %v", err)
	}

	// Read and decrypt the encrypted file
	encPath := filepath.Join("testdata", "enc.yaml")
	actualDecrypted, err := encryptionSvc.Decrypt(encPath, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt enc.yaml: %v", err)
	}

	// Compare (allowing for whitespace differences)
	expectedStr := strings.TrimSpace(string(expectedDecrypted))
	actualStr := strings.TrimSpace(string(actualDecrypted))

	if expectedStr != actualStr {
		t.Errorf("Decrypted content doesn't match dec.yaml\nExpected:\n%s\n\nActual:\n%s",
			expectedStr, actualStr)
	}
}

// TestEditMultipleSecretFields tests editing multiple encrypted fields
func TestEditMultipleSecretFields(t *testing.T) {
	encryptionSvc := encryption.NewSopsAgeDecryptStrategy()

	// Create a test secret with multiple data fields
	secretYAML := `apiVersion: v1
data:
  username: admin
  password: supersecret
  apikey: myapikey123
kind: Secret
metadata:
  name: multi-field-secret
  namespace: default
type: Opaque
`

	// Encrypt the secret
	encrypted, err := encryptionSvc.EncryptData([]byte(secretYAML), testAgePublicKey)
	if err != nil {
		t.Fatalf("Failed to encrypt test secret: %v", err)
	}

	// Decrypt
	decrypted, err := encryptionSvc.DecryptData(encrypted, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Verify all fields present
	decStr := string(decrypted)
	expectedFields := []string{"username: admin", "password: supersecret", "apikey: myapikey123"}
	for _, field := range expectedFields {
		if !strings.Contains(decStr, field) {
			t.Errorf("Decrypted content missing field: %s", field)
		}
	}

	// Simulate editing multiple fields
	modified := strings.Replace(decStr, "admin", "newadmin", 1)
	modified = strings.Replace(modified, "supersecret", "newsupersecret", 1)
	modified = strings.Replace(modified, "myapikey123", "newapikey456", 1)

	// Re-encrypt
	reencrypted, err := encryptionSvc.EncryptData([]byte(modified), testAgePublicKey)
	if err != nil {
		t.Fatalf("Failed to re-encrypt: %v", err)
	}

	// Decrypt and verify changes
	finalDecrypted, err := encryptionSvc.DecryptData(reencrypted, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt re-encrypted data: %v", err)
	}

	finalStr := string(finalDecrypted)
	expectedNewFields := []string{"username: newadmin", "password: newsupersecret", "apikey: newapikey456"}
	for _, field := range expectedNewFields {
		if !strings.Contains(finalStr, field) {
			t.Errorf("Final decrypted content missing updated field: %s", field)
		}
	}
}

// TestEditWithInvalidKey tests that editing with wrong key fails appropriately
func TestEditWithInvalidKey(t *testing.T) {
	encryptionSvc := encryption.NewSopsAgeDecryptStrategy()

	// Read encrypted file
	encryptedPath := filepath.Join("testdata", "enc.yaml")
	encryptedData, err := os.ReadFile(encryptedPath)
	if err != nil {
		t.Fatalf("Failed to read encrypted file: %v", err)
	}

	// Try to decrypt with an invalid key
	invalidKey := "AGE-SECRET-KEY-1AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	_, err = encryptionSvc.DecryptData(encryptedData, invalidKey)
	if err == nil {
		t.Errorf("Expected error when decrypting with invalid key, got nil")
	}
}

// TestPreserveYAMLStructure verifies that YAML structure is preserved after edit cycle
func TestPreserveYAMLStructure(t *testing.T) {
	encryptionSvc := encryption.NewSopsAgeDecryptStrategy()

	// Read and decrypt
	encryptedPath := filepath.Join("testdata", "enc.yaml")
	decrypted, err := encryptionSvc.Decrypt(encryptedPath, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Re-encrypt without changes
	reencrypted, err := encryptionSvc.EncryptData(decrypted, testAgePublicKey)
	if err != nil {
		t.Fatalf("Failed to re-encrypt: %v", err)
	}

	// Decrypt again
	finalDecrypted, err := encryptionSvc.DecryptData(reencrypted, testAgeKey)
	if err != nil {
		t.Fatalf("Failed to decrypt final: %v", err)
	}

	// Check structure elements are preserved
	structureElements := []string{
		"apiVersion:",
		"data:",
		"kind:",
		"metadata:",
		"name:",
		"namespace:",
		"type:",
	}

	finalStr := string(finalDecrypted)
	for _, element := range structureElements {
		if !strings.Contains(finalStr, element) {
			t.Errorf("YAML structure lost element: %s", element)
		}
	}
}
