package domain

type FileService interface {
	CreateTempFile(decrypted []byte) (string, func(), error)
}
