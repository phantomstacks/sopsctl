package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CTX struct {
	PrivateKey string
	Namespace  string
	SecretName string
	KeyName    string
}

func newEmptyCtx() *CTX {
	return &CTX{
		PrivateKey: "",
		Namespace:  "",
		SecretName: "",
		KeyName:    "",
	}
}

type ConfigFile struct {
	StorageMode string
	FilePath    string
	Contexts    map[string]CTX
}

func (c *ConfigFile) SaveStorageMode(mode string) error {
	c.StorageMode = mode
	err := c.SaveConfigFile()
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigFile) GetStorageMode() (string, error) {
	return c.StorageMode, nil
}

func (c *ConfigFile) RemoveCtx(ctxName string) error {
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

func (c *ConfigFile) ListContextsWithKeys() ([]string, error) {
	var result []string
	for ctxName, context := range c.Contexts {
		if context.PrivateKey != "" {
			result = append(result, ctxName)
		}
	}
	return result, nil
}

func (c *ConfigFile) SaveConfigFile() error {
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

func (c *ConfigFile) SetPrivateKey(key string, ctxName string) error {
	ctx := c.getOrCreateCtx(ctxName)
	ctx.PrivateKey = key
	c.Contexts[ctxName] = ctx
	return nil
}

func (c *ConfigFile) getOrCreateCtx(ctxName string) CTX {
	ctx, exists := c.Contexts[ctxName]
	if !exists {
		ctx = *newEmptyCtx()
		c.Contexts[ctxName] = ctx
	}
	return ctx
}

func (c *ConfigFile) GetPrivateKey(ctxName string) (string, error) {
	ctx := c.getOrCreateCtx(ctxName)
	return ctx.PrivateKey, nil
}

func newEmptyConfigFile(filePath string) *ConfigFile {
	err := os.MkdirAll(filepath.Dir(filePath), 0777)
	if err != nil && !os.IsExist(err) {
		return nil
	}
	return &ConfigFile{
		FilePath: filePath,
		Contexts: make(map[string]CTX),
	}
}
