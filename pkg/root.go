package pkg

import (
	"fmt"
	command "phantom-flux/pkg/cmd"
	"phantom-flux/pkg/cmd/key/add"
	"phantom-flux/pkg/cmd/key/list"
	"phantom-flux/pkg/cmd/key/remove"
	storageMode "phantom-flux/pkg/cmd/key/storage"
	"phantom-flux/pkg/cmd/secret/create"
	"phantom-flux/pkg/cmd/secret/decrypt"
	"phantom-flux/pkg/cmd/secret/edit"
	"phantom-flux/pkg/domain"
	"phantom-flux/pkg/services/decoder"
	"phantom-flux/pkg/services/editor"
	"phantom-flux/pkg/services/encryption"
	"phantom-flux/pkg/services/file"
	"phantom-flux/pkg/services/helpers"
	"phantom-flux/pkg/services/key"
	"phantom-flux/pkg/services/storage"

	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

func panicOnError(errs []error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

func GetDigServiceContainer() *dig.Container {
	container := dig.New()

	panicOnError([]error{
		// Core services
		container.Provide(func() domain.KeyStorage {
			return storage.NewLocalUserKeyStorageService()
		}),
		container.Provide(func() domain.SopsKeyManager {
			return key.NewGlobalSopsKeyManager()
		}),
		container.Provide(func() domain.EncryptionService {
			return encryption.NewSopsAgeDecryptStrategy()
		}),
		container.Provide(func() domain.Base64Decoder {
			return decoder.NewBase64Decoder()
		}),
		container.Provide(func() domain.UserEditorService {
			return editor.NewDefaultEditor()
		}),
		container.Provide(func() domain.FileService {
			return file.NewFileService()
		}),

		// Command builders
		container.Provide(func(skm domain.SopsKeyManager) domain.CommandBuilder {
			return list.NewKeyListCmd(skm)
		}, dig.Name(domain.KeyList.ToString())),

		container.Provide(func(skm domain.SopsKeyManager) domain.CommandBuilder {
			return add.NewKeyAddCmd(skm)
		}, dig.Name(domain.KeyAdd.ToString())),

		container.Provide(func(skm domain.SopsKeyManager) domain.CommandBuilder {
			return remove.NewKeyRemoveCmd(skm)
		}, dig.Name(domain.KeyRemove.ToString())),

		container.Provide(func(skm domain.SopsKeyManager, es domain.EncryptionService) domain.CommandBuilder {
			return create.NewSecretCreateCmd(es, skm)
		}, dig.Name(domain.SecretCreate.ToString())),

		container.Provide(func(skm domain.KeyStorage) domain.CommandBuilder {
			return storageMode.NewKeyStorageModeCmd(skm)
		}, dig.Name(domain.KeyStorageMode.ToString())),

		container.Provide(func(
			skm domain.SopsKeyManager,
			encService domain.EncryptionService,
		) domain.CommandBuilder {
			return decrypt.NewSecretDecryptCmd(skm, encService)
		}, dig.Name(domain.SecretDecrypt.ToString())),

		container.Provide(func(
			skm domain.SopsKeyManager,
			encService domain.EncryptionService,
			b64Decoder domain.Base64Decoder,
			editorService domain.UserEditorService,
			fileService domain.FileService,
		) domain.CommandBuilder {
			return edit.NewSecretEditCmd(skm, encService, b64Decoder, editorService, fileService)
		}, dig.Name(domain.SecretEdit.ToString())),

		// CommandFactory
		container.Provide(func(params command.CommandFactoryParams) domain.CommandFactory {
			return command.NewCommandFactory(params)
		}),
	})

	return container
}

func GetCommandBuilder(commandId domain.CommandId) domain.CommandBuilder {
	var builder domain.CommandBuilder = nil
	container := GetDigServiceContainer()
	err := container.Invoke(func(factory domain.CommandFactory) {
		builder = factory.GetCommandBuilder(commandId)
		if builder == nil {
			panic(fmt.Errorf("command builder not found for command: %s", commandId.ToString()))
		}
	})
	if err != nil {
		panic(err)
	}
	return builder
}
func InitCobraCommand(commandId domain.CommandId, cmd *cobra.Command) {
	builder := GetCommandBuilder(commandId)
	builder.InitCmd(cmd)
}

func ExecuteCobraCommand(commandId domain.CommandId, cmd *cobra.Command, args []string) {
	executor, err := GetCommandBuilder(commandId).UseOptions(cmd, args)
	if err != nil {
		helpers.PrintError("failed", err)

		helpText := "To get help, run:\n\n"
		helpText += fmt.Sprintf("  %s --help\n", cmd.CommandPath())
		fmt.Println(helpText)
		return
	}
	result, err := executor.Execute()
	if err != nil {
		helpers.PrintError("Failed to execute command: %v", err)
		return
	}
	fmt.Println(result)
}
