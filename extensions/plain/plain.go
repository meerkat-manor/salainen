package plain

import (
	"fmt"

	"github.com/meerkat-manor/salainen"
)

type f struct{}

const providerName = "plain"

func (sl *f) Init(custom interface{}) error {

	return nil
}

func (sl *f) Put(path, val string) error {

	return fmt.Errorf("plain passwords (clear text) cannot be set")
}

func (sl *f) Get(path string) (string, error) {
	return path, nil
}

func (sl *f) Help() {
	fmt.Printf("Plain help\n\n")
	fmt.Printf("A plain text value can be used, but this does not provide\n")
	fmt.Printf("any level of secret security as the path value is the\n")
	fmt.Printf("the secret.  You cannot save / set a plain text secret.\n")
	fmt.Printf("\n")
	fmt.Printf("For more information please see %s/extensions/plain/ \n", salainen.SourceForgeURL)
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
