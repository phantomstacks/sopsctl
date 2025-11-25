package remove

type KeyRemoveCmdOptions struct {
	RemoveAll    bool
	ClusterNames []string
}

func NewKeyRemoveCmdOptions(showSensitive bool, clusterNames []string) *KeyRemoveCmdOptions {
	return &KeyRemoveCmdOptions{RemoveAll: showSensitive, ClusterNames: clusterNames}
}
