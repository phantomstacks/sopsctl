package domain

type CTX struct {
	PrivateKey string
	Namespace  string
	SecretName string
	KeyName    string
}

func NewReferenceCTX(namespace string, secretName string, keyName string) *CTX {
	return &CTX{Namespace: namespace, SecretName: secretName, KeyName: keyName}
}

func NewEmptyCtx() *CTX {
	return &CTX{
		PrivateKey: "",
		Namespace:  "",
		SecretName: "",
		KeyName:    "",
	}
}

type ConfigStorage interface {
	SaveStorageMode(mode string) error
	GetStorageMode() (string, error)
	SaveConfigFile() error
	SetPrivateKey(key string, ctxName string) error
	GetPrivateKey(ctxName string) (string, error)
	ListContextsWithKeys() ([]string, error)
	RemoveCtx(ctxName string) error
	GetCtx(ctxName string) (*CTX, error)
}
