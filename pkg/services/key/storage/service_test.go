package storage

import (
	"testing"
)

func TestNewLocaUserKeyStorage(t *testing.T) {
	uut := NewLocalUserKeyStorageService()
	if uut == nil {
		t.Fatal("expected non-nil LocalUserKeyStorageService")
	}
}

func TestLocalUserKeyStorageService_SavePrivateKey_GetPrivateKey(t *testing.T) {
	uut := NewLocalUserKeyStorageService()
	key := "some-key"
	ctxName := "some-ctx"
	// Act
	err := uut.SavePrivateKey(
		key,
		ctxName,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	uut = NewLocalUserKeyStorageService()

	privateKey, err := uut.GetPrivateKey(ctxName)
	if err != nil {
		return
	}
	if privateKey != key {
		t.Fatalf("expected key %s, got %s", key, privateKey)
	}
}

func TestLocalUserKeyStorageService_SaveMultiplePrivateKey_GetPrivateKey(t *testing.T) {
	uut := NewLocalUserKeyStorageService()
	key := "some-key"
	ctxName := "some-ctx"

	secondKey := "another-key"
	secondCtxName := "another-ctx"
	// Act
	err := uut.SavePrivateKey(
		key,
		ctxName,
	)
	err = uut.SavePrivateKey(
		secondKey,
		secondCtxName,
	)
	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	uut = NewLocalUserKeyStorageService()

	privateKey, _ := uut.GetPrivateKey(ctxName)
	if privateKey != key {
		t.Fatalf("expected key %s, got %s", key, privateKey)
	}
	secondPrivateKey, _ := uut.GetPrivateKey(secondCtxName)
	if secondPrivateKey != secondKey {
		t.Fatalf("expected key %s, got %s", secondKey, secondPrivateKey)
	}
}
