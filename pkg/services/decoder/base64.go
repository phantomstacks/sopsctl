package decoder

import (
	"encoding/base64"
	"fmt"
	"phantom-flux/pkg/domain"
	"strings"

	"sigs.k8s.io/yaml"
)

type Base64Decoder struct {
}

func NewBase64Decoder() *Base64Decoder {
	return &Base64Decoder{}
}

func (e Base64Decoder) EditDecodedFile(secretFile []byte, valueKey string) ([]byte, func([]byte) ([]byte, error), error) {
	secret := &domain.RawSecret{}
	if err := yaml.Unmarshal(secretFile, &secret); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal secretFile data into Secret object: %w", err)
	}
	isNoData := secret.Data == nil || len(secret.Data) == 0 || secret.Data[valueKey] == ""
	if isNoData {
		return nil, nil, fmt.Errorf("did not find data for key %s in secret", valueKey)
	}
	valueData := strings.TrimSpace(secret.Data[valueKey])

	restoreFunc := e.restoreEncodedFile(*secret, valueKey)

	decodedValue, _ := base64.RawStdEncoding.DecodeString(valueData)
	if decodedValue == nil || len(decodedValue) == 0 {
		return nil, nil, fmt.Errorf("failed to decode base64 value for key %s", valueKey)
	}

	return decodedValue, restoreFunc, nil
}

func (e Base64Decoder) restoreEncodedFile(original domain.RawSecret, editedKey string) func([]byte) ([]byte, error) {
	return func(content []byte) ([]byte, error) {
		encodedContent := base64.RawStdEncoding.EncodeToString(content)
		original.Data[editedKey] = encodedContent
		restoredFile, err := yaml.Marshal(original)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal restored secret data: %w", err)
		}
		return restoredFile, nil
	}
}

func (e Base64Decoder) CountDecodedFileEntries(file []byte) (int, error) {
	secret := &domain.RawSecret{}
	if err := yaml.Unmarshal(file, &secret); err != nil {
		return 0, fmt.Errorf("failed to unmarshal secretFile data into Secret object: %w", err)
	}
	if secret.Data == nil || len(secret.Data) == 0 {
		return 0, nil
	}
	return len(secret.Data), nil
}

func (e Base64Decoder) GetDefaultKey(file []byte) (string, error) {
	secret := &domain.RawSecret{}
	if err := yaml.Unmarshal(file, &secret); err != nil {
		return "", fmt.Errorf("failed to unmarshal secretFile data into Secret object: %w", err)
	}
	if secret.Data == nil || len(secret.Data) == 0 {
		return "", fmt.Errorf("no data found in secret")
	}
	if len(secret.Data) > 1 {
		return "", fmt.Errorf("multiple data entries found in secret")
	}
	for key := range secret.Data {
		return key, nil
	}
	return "", fmt.Errorf("no data found in secret")
}
