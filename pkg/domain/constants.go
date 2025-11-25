package domain

type CommandId string

func (receiver CommandId) ToString() string {
	return string(receiver)
}

const (
	SecretEdit    CommandId = "secret-edit"
	SecretDecrypt CommandId = "secret-decrypt"
	KeyAdd        CommandId = "key-add"
	KeyList       CommandId = "key-list"
	KeyRemove     CommandId = "key-remove"
)
