package domain

import "filippo.io/age"

type SopsKeyManager interface {
	GetIdentityCurrentCtx() (age.Identity, error)
	GetPrivateKey(ctxName string) (string, error)
	GetPublicKey(ctxName string) (string, error)
	AddKeyFromCluster(ctxName string, namespace string, secretName string, secretKey string) (string, error)
	ListContextsWithKeys() ([]string, error)
	RemoveKeyForContext(ctx string) error
}
