package salainen

import (
	"errors"
	"fmt"
	"strings"
)

const (
	StorageSeparator byte   = ':'
	SaneDefaults     bool   = true
	ProductName      string = "salainen"
	ProductVersion   string = "v0.0.1"
	SourceForgeURL   string = "https://github.com/meerkat-manor/salainen/tree/main"
)

var (
	ErrInvalidStorage      = errors.New("invalid storage. Maybe it has not been added")
	ErrNoSuchSecret        = errors.New("no such secret")
	ErrAmbigiousSecret     = errors.New("ambigious secret, provide a subpath after ;")
	ErrInvalidSecret       = errors.New("invalid secret data detected")
	ErrInvalidSecretAccess = errors.New("invalid secret access")
	ErrUnableToSave        = errors.New("unable to save secret")
)

type SecretStorage interface {
	Help()
	Init(custom interface{}) error
	Get(path string) (string, error)
	Put(path string, value string) error
}

var SecretStorages = map[string]SecretStorage{}

func Get(key string) (string, error) {

	if key == "" {
		return "", fmt.Errorf("no key provided")
	}

	storage, keypath, err := SearchSecretStorage(key)
	if err != nil {
		return "", err
	}

	return storage.Get(keypath)
}

func Put(key string, value string) error {

	if key == "" {
		return fmt.Errorf("no key provided")
	}

	storage, keypath, err := SearchSecretStorage(key)
	if err != nil {
		return err
	}

	return storage.Put(keypath, value)
}

func AddSecretStorage(name string, secretStorage SecretStorage) {
	SecretStorages[name] = secretStorage
}

func ListSecretStorage() map[string]SecretStorage {
	return SecretStorages
}

func GetSecretStorage(id string) (SecretStorage, error) {

	storage, exists := SecretStorages[id]
	if !exists {
		return nil, ErrInvalidStorage
	}

	return storage, nil
}

func SearchSecretStorage(id string) (SecretStorage, string, error) {

	if id == "" {
		return nil, "", fmt.Errorf("no storage identifier provided")
	}

	// Allow some sane default handling
	if SaneDefaults {
		if strings.HasPrefix(strings.ToLower(id), "$env:") {
			id = id[1:]
		} else {
			if strings.HasPrefix(id, "${env:") && strings.HasSuffix(id, "}") {
				id = id[2 : len(id)-1]
			} else {
				if strings.HasPrefix(id, "${") && strings.HasSuffix(id, "}") {
					id = "env:" + id[2:len(id)-1]
				}
			}
		}
	}

	idx := strings.IndexByte(id, StorageSeparator)
	if idx < 1 {
		return nil, "", ErrInvalidStorage
	}

	storage, exists := SecretStorages[id[:idx]]
	if !exists {
		return nil, "", ErrInvalidStorage
	}

	return storage, id[idx+1:], nil
}
