package env

import (
	"fmt"
	"os"
	"strings"

	"merebox.com/salainen"
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

func (sl *f) Put(path, val string) error {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	os.Setenv(sl.Prefix+fpath, val)
	return nil
}

func (sl *f) Get(path string) (string, error) {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	return os.Getenv(sl.Prefix + fpath), nil
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
