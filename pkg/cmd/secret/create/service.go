package create

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	openapi2 "k8s.io/client-go/openapi"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/create"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/openapi"
	"k8s.io/kubectl/pkg/validation"

	"phantom-flux/pkg/domain"

	"github.com/spf13/cobra"
)

type factory struct {
}

func newFactory() *factory {
	return &factory{}
}

func (f factory) ToRESTConfig() (*rest.Config, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) ToRESTMapper() (meta.RESTMapper, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	//TODO implement me
	panic("implement me")
}

func (f factory) DynamicClient() (dynamic.Interface, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) KubernetesClientSet() (*kubernetes.Clientset, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) RESTClient() (*rest.RESTClient, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) NewBuilder() *resource.Builder {
	//TODO implement me
	panic("implement me")
}

func (f factory) ClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) UnstructuredClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) Validator(validationDirective string) (validation.Schema, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) OpenAPISchema() (openapi.Resources, error) {
	//TODO implement me
	panic("implement me")
}

func (f factory) OpenAPIV3Client() (openapi2.Client, error) {
	//TODO implement me
	panic("implement me")
}

type SecretCreateCmd struct {
	kubeCreateSecretCmd *cobra.Command
}

func NewSecretCreateCmd() *SecretCreateCmd {
	return &SecretCreateCmd{}
}

func (s SecretCreateCmd) UseOptions(cmd *cobra.Command, args []string) (domain.CommandExecutor, error) {

	return s, nil
}

var o = create.NewSecretOptions(*ioStreams)

func (s SecretCreateCmd) InitCmd(cmd *cobra.Command) {

	util.AddApplyAnnotationFlags(cmd)
	util.AddValidateFlags(cmd)
	util.AddDryRunFlag(cmd)

	cmd.Flags().StringSliceVar(&o.FileSources, "from-file", o.FileSources, "Key files can be specified using their file path, in which case a default name will be given to them, or optionally with a name and file path, in which case the given name will be used.  Specifying a directory will iterate each named file in the directory that is a valid secret key.")
	cmd.Flags().StringArrayVar(&o.LiteralSources, "from-literal", o.LiteralSources, "Specify a key and literal value to insert in secret (i.e. mykey=somevalue)")
	cmd.Flags().StringSliceVar(&o.EnvFileSources, "from-env-file", o.EnvFileSources, "Specify the path to a file to read lines of key=val pairs to create a secret.")
	cmd.Flags().StringVar(&o.Type, "type", o.Type, i18n.T("The type of secret to create"))
	cmd.Flags().BoolVar(&o.AppendHash, "append-hash", o.AppendHash, "Append a hash of the secret to its name.")
}

func (s SecretCreateCmd) Execute() (string, error) {
	return "", nil
}
