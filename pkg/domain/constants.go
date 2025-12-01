package domain

type CommandId string

func (receiver CommandId) ToString() string {
	return string(receiver)
}

const (
	SecretEdit     CommandId = "secret-edit"
	SecretDecrypt  CommandId = "secret-decrypt"
	KeyAdd         CommandId = "key-add"
	KeyList        CommandId = "key-list"
	KeyRemove      CommandId = "key-remove"
	KeyStorageMode CommandId = "key-storage-mode"
)

type StorageMode string

const (
	LocalStorageMode     StorageMode = "local"
	InClusterStorageMode StorageMode = "cluster"
)

func (sm StorageMode) IsValid() bool {
	return sm == LocalStorageMode || sm == InClusterStorageMode
}

func (sm StorageMode) ToString() string {
	return string(sm)
}
