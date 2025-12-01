package key

import (
	"phantom-flux/pkg/domain"
	"phantom-flux/pkg/services/helpers"
	"phantom-flux/pkg/services/storage"

	"filippo.io/age"
	"github.com/fatih/color"
)

type GlobalSopsKeyManager struct {
	storage    domain.KeyStorage
	currentCtx string
}

func (g GlobalSopsKeyManager) GetIdentityCurrentCtx() (age.Identity, error) {
	privateKey, err := g.storage.GetPrivateKey(g.currentCtx)
	if err != nil {
		return nil, err
	}
	identity, err := age.ParseX25519Identity(privateKey)
	if err != nil {
		return nil, err
	}
	return identity, nil
}

func (g GlobalSopsKeyManager) RemoveKeyForContext(ctx string) error {
	err := g.storage.RemoveKeyForContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

// SavePrivateKey saves the private key for the current context unless the storage mode is InCluster.
func (g GlobalSopsKeyManager) SavePrivateKey(key string) error {
	inClusterStorageMode, err := g.isInClusterStorageMode()
	if err != nil {
		return err
	}
	if inClusterStorageMode {
		return nil
	}

	err = g.storage.SavePrivateKey(g.currentCtx, key)
	if err != nil {
		return err
	}
	return nil
}

func (g GlobalSopsKeyManager) isInClusterStorageMode() (bool, error) {
	mode, err := g.storage.GetStorageMode()
	if err != nil {
		return false, err
	}
	if mode == domain.InClusterStorageMode {
		return true, nil
	}
	return false, nil
}

func (g GlobalSopsKeyManager) GetPublicKey(ctxName string) (string, error) {
	inClusterStorageMode, err := g.isInClusterStorageMode()
	if err != nil {
		return "", err
	}
	if inClusterStorageMode {
		identity, err := g.getIdentityFromCluster(ctxName)
		if err != nil {
			return "", err
		}
		return identity.Recipient().String(), nil
	}
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

func (g GlobalSopsKeyManager) getIdentityFromCluster(ctxName string) (*age.X25519Identity, error) {
	ctx, err := g.storage.GetCtx(ctxName)
	if err != nil {
		return nil, err
	}
	strategy, err := createClusterKeyGetterStrategy(ctxName, ctx.Namespace, ctx.SecretName, ctx.KeyName)
	if err != nil {
		return nil, err
	}
	privateKey, err := strategy.Key()
	if err != nil {
		return nil, err
	}
	identity, err := age.ParseX25519Identity(privateKey)
	if err != nil {
		return nil, err
	}
	return identity, nil
}

func (g GlobalSopsKeyManager) ListContextsWithKeys() ([]string, error) {
	return g.storage.ListContextsWithKeys()
}

func (g GlobalSopsKeyManager) GetPrivateKey(ctxName string) (string, error) {
	isInClusterStorageMode, err := g.isInClusterStorageMode()
	if err != nil {
		return "", err
	}
	if isInClusterStorageMode {
		identity, err := g.getIdentityFromCluster(ctxName)
		if err != nil {
			return "", err
		}
		return identity.String(), nil
	}

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

func (g GlobalSopsKeyManager) AddKeyFromCluster(ctxName string, namespace string, secretName string, secretKey string) (string, error) {
	isInClusterStorageMode, err := g.isInClusterStorageMode()
	if err != nil {
		return "", err
	}
	if isInClusterStorageMode {
		err := g.storage.SaveCtxReference(ctxName, namespace, secretName, secretKey)
		if err != nil {
			return "", err
		}
	}
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
		storage: *localUserKeyStorageService,
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
