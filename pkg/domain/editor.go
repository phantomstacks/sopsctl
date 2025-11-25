package domain

import "io"

type UserEditorService interface {
	EditFile(filename string) ([]byte, error)
	EditTempFile(prefix, suffix string, content []byte) ([]byte, func(), error)
	EditStream(prefix, suffix string, r io.Reader) ([]byte, error)
	EditFileWithPostEditCallback(filename string, postEditCallback func([]byte) ([]byte, error)) ([]byte, error)
}
