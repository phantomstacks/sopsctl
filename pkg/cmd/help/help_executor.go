package help

import "github.com/spf13/cobra"

type HelpExecutor struct {
	cmd *cobra.Command
}

func NewHelpExecutor(cmd *cobra.Command) *HelpExecutor {
	return &HelpExecutor{cmd: cmd}
}

func (h HelpExecutor) Execute() (string, error) {
	h.cmd.Help()
	return "", nil
}
