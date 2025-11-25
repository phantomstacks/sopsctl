package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/util/homedir"
)

type ctx struct {
	PrivateKey string
}

func newEmptyCtx() *ctx {
	return &ctx{
		PrivateKey: "",
	}
}

type configFile struct {
	FilePath string
	Contexts map[string]ctx
}

func (c *configFile) RemoveCtx(ctxName string) error {
	_, exists := c.Contexts[ctxName]
	if exists {
		delete(c.Contexts, ctxName)
	}
	if !exists {
		return fmt.Errorf("context %s does not exist", ctxName)
	}
	err := c.SaveConfigFile()
	if err != nil {
		return err
	}
	return nil
}

func (c *configFile) ListContextsWithKeys() ([]string, error) {
	var result []string
	for ctxName, context := range c.Contexts {
		if context.PrivateKey != "" {
			result = append(result, ctxName)
		}
	}
	return result, nil
}

func (c *configFile) SaveConfigFile() error {
	content, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(c.FilePath, content, 0644)
	if err != nil {
		return err
	}
	return err
}

func (c *configFile) SetPrivateKey(key string, ctxName string) error {
	ctx := c.getOrCreateCtx(ctxName)
	ctx.PrivateKey = key
	c.Contexts[ctxName] = ctx
	return nil
}

func (c *configFile) getOrCreateCtx(ctxName string) ctx {
	ctx, exists := c.Contexts[ctxName]
	if !exists {
		ctx = *newEmptyCtx()
		c.Contexts[ctxName] = ctx
	}
	return ctx
}

func (c *configFile) GetPrivateKey(ctxName string) (string, error) {
	ctx := c.getOrCreateCtx(ctxName)
	return ctx.PrivateKey, nil
}

func newEmptyConfigFile(filePath string) *configFile {
	err := os.MkdirAll(filepath.Dir(filePath), 0777)
	if err != nil && !os.IsExist(err) {
		return nil
	}
	return &configFile{
		FilePath: filePath,
		Contexts: make(map[string]ctx),
	}
}

type configStorage interface {
	SaveConfigFile() error
	SetPrivateKey(key string, ctxName string) error
	GetPrivateKey(ctxName string) (string, error)
	ListContextsWithKeys() ([]string, error)
	RemoveCtx(ctxName string) error
}

type LocalUserKeyStorageService struct {
	fileName string
}

func (l LocalUserKeyStorageService) RemoveKeyForContext(ctx string) error {
	config := l.readConfigFromFileOrEmpty()
	err := config.RemoveCtx(ctx)
	if err != nil {
		return err
	}
	return nil
}

func NewLocalUserKeyStorageService() *LocalUserKeyStorageService {
	l := &LocalUserKeyStorageService{}
	l.fileName = "pflux-config.yaml"
	return l
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

func (l LocalUserKeyStorageService) readConfigFromFileOrEmpty() configStorage {
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
	absFilePath := hd + "/.pflux/" + l.fileName
	return absFilePath, nil
}
