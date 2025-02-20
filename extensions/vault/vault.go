package vault

import (
	"context"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/meerkat-manor/salainen"
)

const (
	DefaultVault   = "secret"
	DefaultElement = "password"
)

var providerName = "vault"

type f struct {
	client      *vault.Client
	address     string
	token       string
	vault       string
	elementName string
}

func (sl *f) Init(custom interface{}) error {

	if custom != nil {
		settings := custom.(map[string]string)

		value, exists := settings["ApiUrl"]
		if exists && value != "" {
			sl.address = value
		}

		value, exists = settings["AccessToken"]
		if exists && value != "" {
			sl.token = value
		}

		value, exists = settings["Vault"]
		if exists && value != "" {
			sl.vault = value
		}
		value, exists = settings["ElementName"]
		if exists && value != "" {
			sl.elementName = value
		}

	}

	if sl.address == "" {
		return fmt.Errorf("configuration value for 'ApiUrl' missing")
	}

	if sl.token == "" {
		return fmt.Errorf("configuration value for 'AccessToken' missing")
	}

	if sl.vault == "" {
		sl.vault = DefaultVault
	}

	if sl.elementName == "" {
		sl.elementName = DefaultElement
	}

	if strings.HasPrefix(sl.token, (providerName + ":")) {
		return fmt.Errorf("error fetching %s access password with looping detected", providerName)
	}

	password, errS := salainen.Get(sl.token)
	if errS != nil || password == "" {
		return fmt.Errorf("error fetching %s access password.  More information: %v", providerName, errS)
	}

	sl.token = password
	config := vault.DefaultConfig()
	config.Address = sl.address

	var err error
	sl.client, err = vault.NewClient(config)
	if err != nil {
		return fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	return nil
}

func (sl *f) Put(path, val string) error {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	sl.client.SetToken(sl.token)

	secretData := map[string]interface{}{
		sl.elementName: val,
	}

	ctx := context.Background()

	_, err := sl.client.KVv2(sl.vault).Put(ctx, fpath, secretData)
	if err != nil {
		return fmt.Errorf("unable to write secret: %v", err)
	}

	return nil
}

func (sl *f) Get(path string) (string, error) {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	ctx := context.Background()

	secret, err := sl.client.KVv2(sl.vault).Get(ctx, fpath)
	if err != nil {
		return "", fmt.Errorf("unable to read secret: %v", err)
	}

	value, ok := secret.Data[sl.elementName].(string)
	if !ok {
		return "", fmt.Errorf("value type assertion failed: %v", secret.Data[sl.elementName])
	}

	return value, nil
}

func (sl *f) Help() {
	fmt.Printf("Vault help\n\n")
	fmt.Printf("HashiCorp can be used as a secret provider by using the prefix\n")
	fmt.Printf("'vault:' followed by the key to the secret path in the Vault\n")
	fmt.Printf("\n")
	fmt.Printf("The Vault API URL scheme, host and port needs to be configured\n")
	fmt.Printf("An access token is also mandatory and this could be sourced \n")
	fmt.Printf("from another provider in salainen, such as 'keyring'\n")
	fmt.Printf("\n")
	fmt.Printf("For more information please see %s/extensions/vault/ \n", salainen.SourceForgeURL)
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
