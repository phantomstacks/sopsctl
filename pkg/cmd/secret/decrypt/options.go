package decrypt

type SecretDecryptOptions struct {
	FilePath string
	Cluster  string
}

func NewSecretDecryptOptions(filePath string, cluster string) *SecretDecryptOptions {
	return &SecretDecryptOptions{FilePath: filePath, Cluster: cluster}
}
