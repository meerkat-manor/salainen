package vault

import (
	"context"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"merebox.com/salainen"
)

type f struct {
	client  *vault.Client
	address string
	token   string
}

type CustomConfig struct {
	Address string
	Token   string
}

func (v *f) Init(custom interface{}) error {

	if custom != nil {
		settings := custom.(map[string]interface{})

		value, exists := settings["address"]
		if exists && value.(string) != "" {
			v.address = value.(string)
		}

		value, exists = settings["token"]
		if exists && value.(string) != "" {
			v.token = value.(string)
		}
	}

	config := vault.DefaultConfig()
	config.Address = v.address

	var err error
	v.client, err = vault.NewClient(config)
	if err != nil {
		return fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	return nil
}

func (v *f) Put(path, val string) error {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	v.client.SetToken(v.token)

	// TODO
	secretData := map[string]interface{}{
		"password": val,
	}

	ctx := context.Background()

	_, err := v.client.KVv2(fpath).Put(ctx, "my-secret-password", secretData)
	if err != nil {
		return fmt.Errorf("unable to write secret: %v", err)
	}

	return nil
}

func (v *f) Get(path string) (string, error) {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	ctx := context.Background()

	secret, err := v.client.KVv2(fpath).Get(ctx, "my-secret-password")
	if err != nil {
		return "", fmt.Errorf("unable to read secret: %v", err)
	}

	value, ok := secret.Data["password"].(string)
	if !ok {
		return "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
	}

	return value, nil
}

func (sl *f) Help() {
	fmt.Printf("Vault help\n")
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
	salainen.AddSecretStorage("vault", storage)

	return nil
}
