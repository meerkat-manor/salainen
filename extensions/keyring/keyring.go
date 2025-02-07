package keyring

import (
	"fmt"
	"strings"

	"github.com/meerkat-manor/salainen"
	"github.com/zalando/go-keyring"
)

type f struct {
	Service string
}

const providerName = "keyring"

func (sl *f) Init(custom interface{}) error {

	if custom != nil {
		settings := custom.(map[string]interface{})
		value, exists := settings["Service"]
		if exists && value.(string) != "" {
			sl.Service = value.(string)
			if sl.Service == "{{.ProductName}}" {
				sl.Service = salainen.ProductName
			}
		}
	}

	return nil
}

func (sl *f) Put(path, value string) error {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	return keyring.Set(sl.Service, fpath, value)
}

func (sl *f) Get(path string) (string, error) {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	value, err := keyring.Get(sl.Service, fpath)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("failed to find value")
	}

	return value, nil
}

func (sl *f) Help() {
	fmt.Printf("Keyring help\n")
}

func New(config string, custom interface{}) (salainen.SecretStorage, error) {

	storage := f{}
	err := storage.Init(custom)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

func Register(config string, custom interface{}) error {

	storage, err := New(config, custom)
	if err != nil {
		return err
	}
	salainen.AddSecretStorage(providerName, storage)

	return nil
}
