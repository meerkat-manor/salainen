package salainen

import (
	"fmt"
	"strings"
)

const (
	ProviderSeparator string = ":"
	SaneDefaults      bool   = true
	ProductName       string = "salainen"
	ProductVersion    string = "v0.0.8"
	SourceForgeURL    string = "https://github.com/meerkat-manor/salainen/tree/main"

	MaxProviderLength int = 30
	MaxKeyPathLength  int = 250
	MaxValueLength    int = 600
)

var (
	ErrInvalidProvider     = fmt.Errorf("invalid provider. Maybe it has not been added")
	ErrNoSuchSecret        = fmt.Errorf("no such secret")
	ErrAmbiguousSecret     = fmt.Errorf("ambiguous secret, provide a subpath after %s", ProviderSeparator)
	ErrInvalidSecret       = fmt.Errorf("invalid secret data detected")
	ErrInvalidSecretAccess = fmt.Errorf("invalid secret access")
	ErrUnableToSave        = fmt.Errorf("unable to save secret")
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

	if len(value) < 1 {
		return fmt.Errorf("no value provided")
	}

	if len(value) > MaxValueLength {
		return fmt.Errorf("maximum value length exceeded (%d)", MaxValueLength)
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
		return nil, ErrInvalidProvider
	}

	return storage, nil
}

func SearchSecretStorage(id string) (SecretStorage, string, error) {

	if id == "" {
		return nil, "", fmt.Errorf("no storage identifier provided")
	}

	maxLen := MaxKeyPathLength + MaxProviderLength + 1
	if len(id) > maxLen {
		return nil, "", fmt.Errorf("maximum key length exceeded (%d)", maxLen)
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

	idx := strings.Index(id, ProviderSeparator)
	if idx < 1 {
		return nil, "", ErrInvalidProvider
	}

	storage, exists := SecretStorages[id[:idx]]
	if !exists {
		return nil, "", ErrInvalidProvider
	}

	return storage, id[idx+1:], nil
}
