package domain

type Base64Decoder interface {
	EditDecodedFile(secretFile []byte, valueKey string) ([]byte, func([]byte) ([]byte, error), error)
	CountDecodedFileEntries(file []byte) (int, error)
	GetDefaultKey(file []byte) (string, error)
}
