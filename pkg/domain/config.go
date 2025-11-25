package domain

type ConfigStorage interface {
	SaveStorageMode(mode string) error
	GetStorageMode() (string, error)
	SaveConfigFile() error
	SetPrivateKey(key string, ctxName string) error
	GetPrivateKey(ctxName string) (string, error)
	ListContextsWithKeys() ([]string, error)
	RemoveCtx(ctxName string) error
}
