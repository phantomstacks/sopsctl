package list

type KeyListCmdOptions struct {
	ShowSensitive bool
}

func NewKeyListCmdOptions(showSensitive bool) *KeyListCmdOptions {
	return &KeyListCmdOptions{ShowSensitive: showSensitive}
}
