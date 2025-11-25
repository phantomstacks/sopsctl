package domain

import (
	"github.com/spf13/cobra"
)

type CommandExecutor interface {
	Execute() (string, error)
}

type CommandBuilder interface {
	UseOptions(cmd *cobra.Command, args []string) (CommandExecutor, error)
	InitCmd(cmd *cobra.Command)
}

type CommandFactory interface {
	GetCommandBuilder(cmd CommandId) CommandBuilder
}
