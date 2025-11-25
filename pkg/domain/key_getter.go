package domain

type KeyStrategy interface {
	Key() (string, error)
}
