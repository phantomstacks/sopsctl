package domain

type KeyStorage interface {
	GetCtx(ctxName string) (*CTX, error)
	SetStorageMode(mode StorageMode) error
	GetStorageMode() (StorageMode, error)
	SavePrivateKey(key string, ctxName string) error
	GetPrivateKey(ctxName string) (string, error)
	ListContextsWithKeys() ([]string, error)
	RemoveKeyForContext(ctx string) error
	SaveCtxReference(ctxName string, namespace string, secretName string, key string) error
}
