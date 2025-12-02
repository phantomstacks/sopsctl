package create

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"phantom-flux/pkg/services/utils"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	kubectlcreate "k8s.io/kubectl/pkg/cmd/create"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/hash"
	"k8s.io/kubectl/pkg/util/i18n"

	"phantom-flux/pkg/domain"

	"github.com/spf13/cobra"
)

// SecretCreateCmd wraps kubectl's secret creation logic with encryption
type SecretCreateCmd struct {
	// Options mirror kubectl's CreateSecretOptions but without the k8s client
	Name           string
	Type           string
	FileSources    []string
	LiteralSources []string
	EnvFileSources []string
	AppendHash     bool
	Namespace      string
	Cluster        string

	// IOStreams for output
	IOStreams         genericiooptions.IOStreams
	encryptionService domain.EncryptionService
	sopsKeyManager    domain.SopsKeyManager
}

func NewSecretCreateCmd(es domain.EncryptionService, skm domain.SopsKeyManager) *SecretCreateCmd {
	return &SecretCreateCmd{
		encryptionService: es,
		sopsKeyManager:    skm,
	}
}

func (s *SecretCreateCmd) UseOptions(cmd *cobra.Command, args []string) (domain.CommandExecutor, error) {
	s.IOStreams = genericiooptions.IOStreams{
		In:     cmd.InOrStdin(),
		Out:    cmd.OutOrStdout(),
		ErrOut: cmd.ErrOrStderr(),
	}
	flags, err := utils.UseGlobalFlags(cmd)
	if err != nil {
		return nil, err
	}
	s.Cluster = flags.Cluster

	// Parse the secret name from args
	name, err := kubectlcreate.NameFromCommandArgs(cmd, args)
	if err != nil {
		return nil, err
	}
	s.Name = name

	// Get namespace from flags or use default
	namespace, _ := cmd.Flags().GetString("namespace")
	if namespace == "" {
		namespace = "default"
	}
	s.Namespace = namespace

	return s, nil
}

func (s *SecretCreateCmd) InitCmd(cmd *cobra.Command) {
	// Add all the standard kubectl secret generic flags
	cmd.Flags().StringSliceVar(&s.FileSources, "from-file", s.FileSources, "Key files can be specified using their file path, in which case a default name will be given to them, or optionally with a name and file path, in which case the given name will be used.  Specifying a directory will iterate each named file in the directory that is a valid secret key.")
	cmd.Flags().StringArrayVar(&s.LiteralSources, "from-literal", s.LiteralSources, "Specify a key and literal value to insert in secret (i.e. mykey=somevalue)")
	cmd.Flags().StringSliceVar(&s.EnvFileSources, "from-env-file", s.EnvFileSources, "Specify the path to a file to read lines of key=val pairs to create a secret.")
	cmd.Flags().StringVar(&s.Type, "type", s.Type, i18n.T("The type of secret to create"))
	cmd.Flags().BoolVar(&s.AppendHash, "append-hash", s.AppendHash, "Append a hash of the secret to its name.")
	cmd.Flags().StringVarP(&s.Namespace, "namespace", "n", s.Namespace, "Namespace for the secret")
}

func (s *SecretCreateCmd) Execute() (string, error) {
	// Validate inputs
	if err := s.Validate(); err != nil {
		return "", err
	}

	// Create the secret object using kubectl's logic
	secret, err := s.createSecret()
	if err != nil {
		return "", fmt.Errorf("failed to create secret: %w", err)
	}

	publicKey, err := s.sopsKeyManager.GetPublicKey(s.Cluster)
	if err != nil {
		return "", err
	}

	// Convert secret to YAML using Kubernetes printer (proper formatting with capitalized fields)
	secretBytes, err := s.marshalSecretToYAML(secret)
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret: %w", err)
	}

	encryptedSecret, err := s.encryptionService.EncryptData(secretBytes, publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	return string(encryptedSecret), nil
}

// marshalSecretToYAML uses the Kubernetes YAML printer to properly format the secret
// This ensures correct field names (apiVersion, not apiversion) like kubectl does
func (s *SecretCreateCmd) marshalSecretToYAML(secret *corev1.Secret) ([]byte, error) {
	// Create a YAML printer with the Kubernetes scheme
	printer := printers.NewTypeSetter(scheme.Scheme).ToPrinter(&printers.YAMLPrinter{})

	// Print to a buffer
	var buf bytes.Buffer
	if err := printer.PrintObj(secret, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *SecretCreateCmd) Validate() error {
	if len(s.Name) == 0 {
		return fmt.Errorf("name must be specified")
	}
	if len(s.EnvFileSources) > 0 && (len(s.FileSources) > 0 || len(s.LiteralSources) > 0) {
		return fmt.Errorf("from-env-file cannot be combined with from-file or from-literal")
	}
	return nil
}

// createSecret is based on kubectl's CreateSecretOptions.createSecret()
// This creates the secret object without interacting with the Kubernetes API
func (s *SecretCreateCmd) createSecret() (*corev1.Secret, error) {
	secret := newSecretObj(s.Name, s.Namespace, corev1.SecretType(s.Type))

	if len(s.LiteralSources) > 0 {
		if err := handleSecretFromLiteralSources(secret, s.LiteralSources); err != nil {
			return nil, err
		}
	}
	if len(s.FileSources) > 0 {
		if err := handleSecretFromFileSources(secret, s.FileSources); err != nil {
			return nil, err
		}
	}
	if len(s.EnvFileSources) > 0 {
		if err := handleSecretFromEnvFileSources(secret, s.EnvFileSources); err != nil {
			return nil, err
		}
	}
	if s.AppendHash {
		hashValue, err := hash.SecretHash(secret)
		if err != nil {
			return nil, err
		}
		secret.Name = fmt.Sprintf("%s-%s", secret.Name, hashValue)
	}

	return secret, nil
}

// Helper functions copied from kubectl's create_secret.go
// These are the core functions that build the secret from different sources

func newSecretObj(name, namespace string, secretType corev1.SecretType) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: secretType,
		Data: map[string][]byte{},
	}
}

func handleSecretFromLiteralSources(secret *corev1.Secret, literalSources []string) error {
	for _, literalSource := range literalSources {
		keyName, value, err := util.ParseLiteralSource(literalSource)
		if err != nil {
			return err
		}
		if err = addKeyFromLiteralToSecret(secret, keyName, []byte(value)); err != nil {
			return err
		}
	}
	return nil
}

func handleSecretFromFileSources(secret *corev1.Secret, fileSources []string) error {
	for _, fileSource := range fileSources {
		keyName, filePath, err := util.ParseFileSource(fileSource)
		if err != nil {
			return err
		}
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			switch err := err.(type) {
			case *os.PathError:
				return fmt.Errorf("error reading %s: %v", filePath, err.Err)
			default:
				return fmt.Errorf("error reading %s: %v", filePath, err)
			}
		}
		// if the filePath is a directory
		if fileInfo.IsDir() {
			if strings.Contains(fileSource, "=") {
				return fmt.Errorf("cannot give a key name for a directory path")
			}
			fileList, err := os.ReadDir(filePath)
			if err != nil {
				return fmt.Errorf("error listing files in %s: %v", filePath, err)
			}
			for _, item := range fileList {
				itemPath := filepath.Join(filePath, item.Name())
				if item.Type().IsRegular() {
					keyName = item.Name()
					if err := addKeyFromFileToSecret(secret, keyName, itemPath); err != nil {
						return err
					}
				}
			}
		} else {
			if err := addKeyFromFileToSecret(secret, keyName, filePath); err != nil {
				return err
			}
		}
	}
	return nil
}

func handleSecretFromEnvFileSources(secret *corev1.Secret, envFileSources []string) error {
	for _, envFileSource := range envFileSources {
		info, err := os.Stat(envFileSource)
		if err != nil {
			switch err := err.(type) {
			case *os.PathError:
				return fmt.Errorf("error reading %s: %v", envFileSource, err.Err)
			default:
				return fmt.Errorf("error reading %s: %v", envFileSource, err)
			}
		}
		if info.IsDir() {
			return fmt.Errorf("env secret file cannot be a directory")
		}
		err = cmdutil.AddFromEnvFile(envFileSource, func(key, value string) error {
			return addKeyFromLiteralToSecret(secret, key, []byte(value))
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func addKeyFromFileToSecret(secret *corev1.Secret, keyName, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return addKeyFromLiteralToSecret(secret, keyName, data)
}

func addKeyFromLiteralToSecret(secret *corev1.Secret, keyName string, data []byte) error {
	if errs := validation.IsConfigMapKey(keyName); len(errs) != 0 {
		return fmt.Errorf("%q is not valid key name for a Secret %s", keyName, strings.Join(errs, ";"))
	}
	if _, entryExists := secret.Data[keyName]; entryExists {
		return fmt.Errorf("cannot add key %s, another key by that name already exists", keyName)
	}
	secret.Data[keyName] = data
	return nil
}
