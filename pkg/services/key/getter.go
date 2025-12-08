package key

import (
	"context"
	"fmt"
	"sopsctl/pkg/domain"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type GetFromClusterKeyStrategy struct {
	kubeClient kubernetes.Interface
	namespace  string
	secretName string
	secretKey  string
}

func NewGetFromClusterKeyStrategy(kubeClient kubernetes.Interface, namespace string, secretName string, secretKey string) domain.KeyStrategy {
	return &GetFromClusterKeyStrategy{
		kubeClient: kubeClient,
		namespace:  namespace,
		secretName: secretName,
		secretKey:  secretKey,
	}
}

func (m *GetFromClusterKeyStrategy) Key() (string, error) {
	ctx := context.Background()
	secret, err := m.kubeClient.CoreV1().Secrets(m.namespace).Get(ctx, m.secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	// Extract the key from the secret data
	key, ok := secret.Data[m.secretKey]
	if !ok {
		return "", fmt.Errorf("key not found in secret")
	}

	cleanedKey := getPrivateKeyFromAgeFileContent(string(key))

	return cleanedKey, nil
}

func getPrivateKeyFromAgeFileContent(file string) string {
	// AGE keys should have the format:
	// # created: timestamp
	// # public key: ...
	// AGE-SECRET-KEY-1...

	// Find the index of "AGE-SECRET-KEY"
	index := strings.Index(file, "AGE-SECRET-KEY")
	if index == -1 {
		return "" // Not found
	}

	// Extract from the start of "AGE-SECRET-KEY" to the end of that line
	startOfLine := index

	// Find the end of the line (newline character)
	endIndex := strings.Index(file[index:], "\n")
	if endIndex == -1 {
		// No newline, take rest of string
		return strings.TrimSpace(file[startOfLine:])
	}

	result := strings.TrimSpace(file[startOfLine : startOfLine+endIndex])
	return result
}
