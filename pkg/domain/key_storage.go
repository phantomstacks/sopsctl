package domain

type CtxKeyPair struct {
	PrivateKey string
	PublicKey  string
	CtxName    string
}

type KeyStorage interface {
	SavePrivateKey(key string, ctxName string) error
	GetPrivateKey(ctxName string) (string, error)
	ListContextsWithKeys() ([]string, error)
	RemoveKeyForContext(ctx string) error
}
