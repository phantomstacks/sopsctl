package key

import (
	"phantom-flux/pkg/domain"
	"phantom-flux/pkg/services/helpers"
	"phantom-flux/pkg/services/key/storage"

	"filippo.io/age"
	"github.com/fatih/color"
)

type GlobalSopsKeyManager struct {
	storage    domain.KeyStorage
	currentCtx string
}

func (g GlobalSopsKeyManager) RemoveKeyForContext(ctx string) error {
	err := g.storage.RemoveKeyForContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (g GlobalSopsKeyManager) SavePrivateKey(key string) error {
	err := g.storage.SavePrivateKey(g.currentCtx, key)
	if err != nil {
		return err
	}
	return nil
}

func (g GlobalSopsKeyManager) GetPublicKey(ctxName string) (string, error) {
	key, err := g.storage.GetPrivateKey(ctxName)
	if err != nil {
		return "", err
	}
	identity, err := age.ParseX25519Identity(key)
	if err != nil {
		return "", err
	}
	return identity.Recipient().String(), nil
}

func (g GlobalSopsKeyManager) ListContextsWithKeys() ([]string, error) {
	return g.storage.ListContextsWithKeys()
}

func (g GlobalSopsKeyManager) GetPrivateKey(ctxName string) (string, error) {
	key, err := g.storage.GetPrivateKey(ctxName)
	if err != nil {
		return "", err
	}
	_, err = age.ParseX25519Identity(key)
	if err != nil {
		return "", err
	}
	return key, nil
}

func (g GlobalSopsKeyManager) GetIdentityCurrentCtx() (age.Identity, error) {
	key, err := g.storage.GetPrivateKey(g.currentCtx)
	if err != nil {
		return nil, err
	}
	identity, err := age.ParseX25519Identity(key)
	if err != nil {
		return nil, err
	}
	return identity, nil
}

func (g GlobalSopsKeyManager) AddKeyFromCluster(ctxName string, namespace string, secretName string, secretKey string) (string, error) {
	clusterKeyGetter, err := createClusterKeyGetterStrategy(ctxName, namespace, secretName, secretKey)
	if err != nil {
		return "", err
	}
	if ctxName == "" {
		ctxName, err = helpers.GetCtxNameFromCurrent()
		if err != nil {
			return "", err
		}
	}
	privateKey, err := clusterKeyGetter.Key()
	if err != nil {
		return "", err
	}
	_, err = age.ParseX25519Identity(privateKey)
	if err != nil {
		return "", err
	}
	err = g.storage.SavePrivateKey(privateKey, ctxName)
	if err != nil {
		return "", err
	}
	return "Added sops key from cluster secret" + ": " + color.GreenString(ctxName) + "/" + color.GreenString(namespace) + "/" + color.GreenString(secretName) + ":(" + color.GreenString(secretKey) + ") in local storage", nil
}

func NewGlobalSopsKeyManager() *GlobalSopsKeyManager {
	localUserKeyStorageService := storage.NewLocalUserKeyStorageService()
	return &GlobalSopsKeyManager{
		storage: localUserKeyStorageService,
	}
}

func createClusterKeyGetterStrategy(ctxName string, namespace string, secretName string, secretKey string) (domain.KeyStrategy, error) {
	client, err := helpers.GetKubeClientForContext(ctxName)
	if err != nil {
		return nil, err
	}
	clusterKeyGetter := NewGetFromClusterKeyStrategy(client, namespace, secretName, secretKey)
	return clusterKeyGetter, nil
}
