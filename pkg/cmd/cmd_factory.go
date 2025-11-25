package cmd

import (
	"fmt"
	"phantom-flux/pkg/domain"

	"go.uber.org/dig"
)

type CommandFactoryParams struct {
	dig.In
	KeyAddCmdBuilder        domain.CommandBuilder `name:"key-add"`
	KeyListCmdBuilder       domain.CommandBuilder `name:"key-list"`
	KeyRemoveCmdBuilder     domain.CommandBuilder `name:"key-remove"`
	SecretEditCmdBuilder    domain.CommandBuilder `name:"secret-edit"`
	SecretDecryptCmdBuilder domain.CommandBuilder `name:"secret-decrypt"`
}

type CommandFactory struct {
	keyAddCmdBuilder        domain.CommandBuilder
	keyListCmdBuilder       domain.CommandBuilder
	keyRemoveCmdBuilder     domain.CommandBuilder
	secretEditCmdBuilder    domain.CommandBuilder
	secretDecryptCmdBuilder domain.CommandBuilder
}

func NewCommandFactory(params CommandFactoryParams) *CommandFactory {
	return &CommandFactory{
		keyAddCmdBuilder:        params.KeyAddCmdBuilder,
		keyListCmdBuilder:       params.KeyListCmdBuilder,
		keyRemoveCmdBuilder:     params.KeyRemoveCmdBuilder,
		secretEditCmdBuilder:    params.SecretEditCmdBuilder,
		secretDecryptCmdBuilder: params.SecretDecryptCmdBuilder,
	}
}

func (cf *CommandFactory) GetCommandBuilder(cmd domain.CommandId) domain.CommandBuilder {
	switch cmd {
	case domain.KeyAdd:
		return cf.keyAddCmdBuilder
	case domain.KeyList:
		return cf.keyListCmdBuilder
	case domain.KeyRemove:
		return cf.keyRemoveCmdBuilder
	case domain.SecretEdit:
		return cf.secretEditCmdBuilder
	case domain.SecretDecrypt:
		return cf.secretDecryptCmdBuilder

	default:
		panic(fmt.Errorf("unknown command: %s", cmd))
	}
}
