package wincred

import (
	"fmt"

	"github.com/danieljoos/wincred"
	"github.com/meerkat-manor/salainen"
)

type f struct {
	Prefix string
}

var providerName = "wincred"

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

	if sl.Prefix != "" {
		path = sl.Prefix + path
	}

	cred := wincred.NewGenericCredential(path)
	cred.CredentialBlob = []byte(value)

	err := cred.Write()
	if err != nil {
		return err
	}

	return nil
}

func (sl *f) Get(path string) (string, error) {

	if sl.Prefix != "" {
		path = sl.Prefix + path
	}

	cred, err := wincred.GetGenericCredential(path)
	if err != nil {
		return "", err
	}

	return string(cred.CredentialBlob), nil
}

func (sl *f) Help() {
	fmt.Printf("Microsoft Windows Credential Manager help\n\n")
	fmt.Printf("The Microsoft Windows Credential Manager can be used as a\n")
	fmt.Printf("secret provider by using the prefix 'wincred:' followed\n")
	fmt.Printf("by the key to the secret in the Credential Manager\n")
	fmt.Printf("\n")
	fmt.Printf("As the credentials are tied to the logged in user, this\n")
	fmt.Printf("is one of the most secure providers available with the tool.\n")
	fmt.Printf("The provider is only available on platforms supported by Microsoft\n")
	fmt.Printf("\n")
	fmt.Printf("For more information please see %s/extensions/wincred/ \n", salainen.SourceForgeURL)
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
