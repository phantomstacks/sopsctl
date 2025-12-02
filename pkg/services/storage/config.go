package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"phantom-flux/pkg/domain"

	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	StorageMode string
	FilePath    string
	Contexts    map[string]domain.CTX
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
	c.Contexts[ctxName] = *ctx
	return nil
}

func (c *ConfigFile) getOrCreateCtx(ctxName string) *domain.CTX {
	ctx, exists := c.Contexts[ctxName]
	if !exists {
		ctx = *domain.NewEmptyCtx()
		c.Contexts[ctxName] = ctx
	}
	return &ctx
}

func (c *ConfigFile) GetPrivateKey(ctxName string) (string, error) {
	ctx := c.getOrCreateCtx(ctxName)
	return ctx.PrivateKey, nil
}

func (c *ConfigFile) GetCtx(name string) (*domain.CTX, error) {
	ctx, exists := c.Contexts[name]
	if !exists {
		return nil, fmt.Errorf("context %s does not exist", name)
	}
	return &ctx, nil
}

func (c *ConfigFile) SaveCtx(ctxName string, ctx *domain.CTX) error {
	c.Contexts[ctxName] = *ctx
	err := c.SaveConfigFile()
	if err != nil {
		return err
	}
	return nil
}

func newEmptyConfigFile(filePath string) *ConfigFile {
	err := os.MkdirAll(filepath.Dir(filePath), 0777)
	if err != nil && !os.IsExist(err) {
		return nil
	}
	return &ConfigFile{
		FilePath: filePath,
		Contexts: make(map[string]domain.CTX),
	}
}
