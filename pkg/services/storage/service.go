package storage

import (
	"os"
	"phantom-flux/pkg/domain"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/util/homedir"
)

func NewLocalUserKeyStorageService() *LocalUserKeyStorageService {
	l := &LocalUserKeyStorageService{}
	l.fileName = "sopsctl-config.yaml"
	return l
}

type LocalUserKeyStorageService struct {
	fileName string
}

func (l LocalUserKeyStorageService) SaveCtxReference(ctxName string, namespace string, secretName string, key string) error {
	config := l.readConfigFromFileOrEmpty()
	ctx := domain.NewReferenceCTX(namespace, secretName, key)

	err := config.SaveCtx(ctxName, ctx)
	return err
}

func (l LocalUserKeyStorageService) GetCtx(ctxName string) (*domain.CTX, error) {
	config := l.readConfigFromFileOrEmpty()
	ctx, err := config.GetCtx(ctxName)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func (l LocalUserKeyStorageService) GetStorageMode() (domain.StorageMode, error) {
	config := l.readConfigFromFileOrEmpty()
	mode, err := config.GetStorageMode()
	if err != nil {
		return "", err
	}
	return domain.StorageMode(mode), nil
}

func (l LocalUserKeyStorageService) SetStorageMode(mode domain.StorageMode) error {
	config := l.readConfigFromFileOrEmpty()
	err := config.SaveStorageMode(mode.ToString())
	if err != nil {
		return err
	}
	return nil
}

func (l LocalUserKeyStorageService) RemoveKeyForContext(ctx string) error {
	config := l.readConfigFromFileOrEmpty()
	err := config.RemoveCtx(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (l LocalUserKeyStorageService) ListContextsWithKeys() ([]string, error) {
	config := l.readConfigFromFileOrEmpty()
	return config.ListContextsWithKeys()
}

func (l LocalUserKeyStorageService) SavePrivateKey(key string, ctxName string) error {
	config := l.readConfigFromFileOrEmpty()
	err := config.SetPrivateKey(key, ctxName)
	if err != nil {
		return err
	}
	err = config.SaveConfigFile()
	if err != nil {
		return err
	}
	return nil
}

func (l LocalUserKeyStorageService) GetPrivateKey(ctxName string) (string, error) {
	config := l.readConfigFromFileOrEmpty()
	return config.GetPrivateKey(ctxName)
}

func (l LocalUserKeyStorageService) readConfigFromFileOrEmpty() *ConfigFile {
	file, err := l.getAbsoluteFilePath()
	if err != nil {
		return newEmptyConfigFile(file)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return newEmptyConfigFile(file)
	}

	var config = newEmptyConfigFile(file)
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return newEmptyConfigFile(file)
	}
	return config
}

func (l LocalUserKeyStorageService) getAbsoluteFilePath() (string, error) {
	hd := homedir.HomeDir()
	if hd == "" {
		return "", nil
	}
	absFilePath := hd + "/.sopsctl/" + l.fileName
	return absFilePath, nil
}
