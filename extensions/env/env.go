package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/meerkat-manor/salainen"
)

type f struct {
	Prefix string
}

const providerName = "env"

func (sl *f) Init(custom interface{}) error {

	if custom != nil {
		settings := custom.(map[string]interface{})
		value, exists := settings["Prefix"]
		if exists && value.(string) != "" {
			sl.Prefix = value.(string)
			if sl.Prefix == "{{.ProductName}}" {
				sl.Prefix = salainen.ProductName
			}
		}
	}

	return nil
}

func (sl *f) Put(path, value string) error {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	return os.Setenv(sl.Prefix+fpath, value)
}

func (sl *f) Get(path string) (string, error) {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	value := os.Getenv(sl.Prefix + fpath)

	if value == "" {
		return "", fmt.Errorf("failed to find value")
	}

	return value, nil
}

func (sl *f) Help() {
	fmt.Printf("Environment help\n")
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
