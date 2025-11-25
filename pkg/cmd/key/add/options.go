package add

type KeyAddCmdOptions struct {
	Cluster    string
	Namespace  string
	SecretName string
	SecretKey  string
}

func NewKeyAddCmdOptions(cluster string, namespace string, secretName string, secretKey string) *KeyAddCmdOptions {
	return &KeyAddCmdOptions{Cluster: cluster, Namespace: namespace, SecretName: secretName, SecretKey: secretKey}
}
