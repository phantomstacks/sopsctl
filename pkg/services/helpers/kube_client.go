package helpers

import (
	"fmt"

	"github.com/fatih/color"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubeClientForContext(ctxName string) (kubernetes.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	rawConfig, err := loadingRules.Load()
	if err != nil {
		return nil, err
	}
	if _, ok := rawConfig.Contexts[ctxName]; !ok {
		return nil, fmt.Errorf("context %q not found in kubeconfig", ctxName)
	}

	overrides := &clientcmd.ConfigOverrides{CurrentContext: ctxName}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(restConfig)
}

func GetCtxNameFromCurrent() (string, error) {
	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return "", err
	}
	return config.CurrentContext, nil
}

func PrintError(s string, err error) {
	color.Red("%s: %v", s, err)
}

func RandomString(i int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, i)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
