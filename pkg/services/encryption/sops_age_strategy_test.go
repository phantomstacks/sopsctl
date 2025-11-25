package encryption

import (
	"os"
	"strings"
	"testing"
	"unicode"
)

func TestNewSopsAgeDecryptStrategy(t *testing.T) {
	strategy := NewSopsAgeDecryptStrategy()
	if strategy == nil {
		t.Fatal("expected non-nil strategy")
	}
}

func TestSopsAgeDecryptStrategy_Decrypt_InvalidAgeKey(t *testing.T) {
	// Setup
	strategy := NewSopsAgeDecryptStrategy()

	// Act
	_, err := strategy.Decrypt("test.yaml", "invalid-key")

	// Assert
	noError := err == nil
	if noError {
		t.Error("expected error for invalid age key")
	}
}

func TestSopsAgeDecryptStrategy_Decrypt_FileNotFound(t *testing.T) {
	// Setup
	strategy := NewSopsAgeDecryptStrategy()
	validKey := "AGE-SECRET-KEY-1QQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ"

	// Act
	_, err := strategy.Decrypt("non-existent-file.yaml", validKey)

	// Assert
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestSopsAgeDecryptStrategy_Decrypt_Success(t *testing.T) {

	// Setup
	strategy := NewSopsAgeDecryptStrategy()
	key := "AGE-SECRET-KEY-13ZLWP4WFHQ6VHC2J5YYEUCFKGLZTD3SXQQPEGK3WU2M8FKYC238S7ZKNSV"

	unencrypted, _ := os.ReadFile("./testdata/dec.yaml")
	trimmedUnencrypted := removeWhitespace(string(unencrypted))

	// Act
	decryptedData, _ := strategy.Decrypt("./testdata/enc.yaml", key)

	// Assert
	trimmedDecData := removeWhitespace(string(decryptedData))
	if trimmedDecData != trimmedUnencrypted {
		t.Errorf("decrypted data does not match expected cleartext")
	}

}

func removeWhitespace(file string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, file)
}
