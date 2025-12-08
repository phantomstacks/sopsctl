package decoder

import (
	"encoding/base64"
	"sopsctl/pkg/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBase64Decoder_EditDecodedFile_Success(t *testing.T) {
	decoder := Base64Decoder{}

	// Create a valid Kubernetes secret with a single data key
	// Note: In YAML, the data values are plain text, yaml.Unmarshal converts them to []byte
	const yamlContent = "TestContent"
	secretYAML := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
data:
  config.yaml: ` + base64.RawStdEncoding.EncodeToString([]byte(yamlContent)) + `
type: Opaque`

	// Call EditDecodedFile
	decoded, restoreFunc, err := decoder.EditDecodedFile([]byte(secretYAML), "config.yaml")

	// Assert no error
	require.NoError(t, err)
	assert.NotNil(t, restoreFunc)

	// Assert the decoded content is correct (should be base64 decoded)
	assert.Equal(t, yamlContent, string(decoded))
}

func TestBase64Decoder_EditDecodedFile_NoDataKeys(t *testing.T) {
	decoder := Base64Decoder{}

	// Secret with no data keys
	secretYAML := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
data: {}
type: Opaque`

	// Call EditDecodedFile
	decoded, restoreFunc, err := decoder.EditDecodedFile([]byte(secretYAML), "config.yaml")

	// Assert error occurred
	assert.Error(t, err)
	assert.Nil(t, decoded)
	assert.Nil(t, restoreFunc)
	assert.Contains(t, err.Error(), "did not find data for key")
}

func TestBase64Decoder_RestoreEncodedFile_Success(t *testing.T) {
	decoder := Base64Decoder{}

	// Create a valid secret
	secretYAML := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
data:
  config.yaml: b3JpZ2luYWw6IGNvbnRlbnQ
type: Opaque`

	// Call EditDecodedFile to get the restore function
	_, restoreFunc, err := decoder.EditDecodedFile([]byte(secretYAML), "config.yaml")
	require.NoError(t, err)
	require.NotNil(t, restoreFunc)

	// Modify the content and restore it
	modifiedContent := []byte("modified: content\nnewkey: newvalue")
	restored, err := restoreFunc(modifiedContent)

	// Assert no error
	require.NoError(t, err)
	assert.NotNil(t, restored)

	// Parse the restored YAML and verify the structure
	var restoredSecret = &domain.RawSecret{}
	err = yaml.Unmarshal(restored, &restoredSecret)
	require.NoError(t, err)

	// Verify the secret has the expected structure
	assert.Equal(t, "v1", restoredSecret.APIVersion)
	assert.Equal(t, "Secret", restoredSecret.Kind)

	// Verify the data is base64 encoded
	data := restoredSecret.Data
	encodedValue := data["config.yaml"]

	// Decode and verify the content
	decodedValue, err := base64.RawStdEncoding.DecodeString(encodedValue)
	require.NoError(t, err)
	assert.Equal(t, string(modifiedContent), string(decodedValue))
}

func TestBase64Decoder_RestoreEncodedFile_EmptyContent(t *testing.T) {
	decoder := Base64Decoder{}

	// Create a valid secret
	secretYAML := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
data:
  config.yaml: b3JpZ2luYWw
type: Opaque`

	// Call EditDecodedFile to get the restore function
	_, restoreFunc, err := decoder.EditDecodedFile([]byte(secretYAML), "config.yaml")
	require.NoError(t, err)

	// Restore with empty content
	emptyContent := []byte("")
	restored, err := restoreFunc(emptyContent)

	// Assert no error
	require.NoError(t, err)
	assert.NotNil(t, restored)

	// Verify the restored secret contains empty base64 content
	var restoredSecret map[string]interface{}
	err = yaml.Unmarshal(restored, &restoredSecret)
	require.NoError(t, err)

	data := restoredSecret["data"].(map[string]interface{})
	encodedValue := data["config.yaml"].(string)

	decodedValue, err := base64.RawStdEncoding.DecodeString(encodedValue)
	require.NoError(t, err)
	assert.Equal(t, "", string(decodedValue))
}

func TestBase64Decoder_RestoreEncodedFile_LargeContent(t *testing.T) {
	decoder := Base64Decoder{}

	// Create a valid secret
	secretYAML := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
data:
  config.yaml: c21hbGw
type: Opaque`

	// Call EditDecodedFile to get the restore function
	_, restoreFunc, err := decoder.EditDecodedFile([]byte(secretYAML), "config.yaml")
	require.NoError(t, err)

	// Create large content
	largeContent := make([]byte, 10000)
	for i := range largeContent {
		largeContent[i] = byte('a' + (i % 26))
	}

	// Restore with large content
	restored, err := restoreFunc(largeContent)

	// Assert no error
	require.NoError(t, err)
	assert.NotNil(t, restored)

	// Verify the content is correctly restored
	var restoredSecret map[string]interface{}
	err = yaml.Unmarshal(restored, &restoredSecret)
	require.NoError(t, err)

	data := restoredSecret["data"].(map[string]interface{})
	encodedValue := data["config.yaml"].(string)

	decodedValue, err := base64.RawStdEncoding.DecodeString(encodedValue)
	require.NoError(t, err)
	assert.Equal(t, largeContent, decodedValue)
}

func TestBase64Decoder_RoundTrip(t *testing.T) {
	decoder := Base64Decoder{}

	originalContent := "key: value\nfoo: bar\nnested:\n  item: test"
	secretYAML := `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
data:
  config.yaml: a2V5OiB2YWx1ZQpmb286IGJhcgpuZXN0ZWQ6CiAgaXRlbTogdGVzdA
type: Opaque`

	// Decode
	decoded, restoreFunc, err := decoder.EditDecodedFile([]byte(secretYAML), "config.yaml")
	require.NoError(t, err)
	assert.Equal(t, originalContent, string(decoded))

	// Modify
	modifiedContent := decoded
	modifiedContent = append(modifiedContent, []byte("\nadded: field")...)

	// Restore
	restored, err := restoreFunc(modifiedContent)
	require.NoError(t, err)

	// Decode again to verify
	decoded2, _, err := decoder.EditDecodedFile(restored, "config.yaml")
	require.NoError(t, err)
	assert.Equal(t, string(modifiedContent), string(decoded2))
}
