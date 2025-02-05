package promptsec

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/meerkat-manor/salainen"
)

type f struct {
	Prompt string
}

const providerName = "prompt"

func (sl *f) Init(custom interface{}) error {

	if sl.Prompt == "" {
		sl.Prompt = "Please input the secret"
	}

	return nil
}

func (sl *f) Put(path, val string) error {

	return fmt.Errorf("prompt passwords cannot be set")
}

func (sl *f) Get(path string) (string, error) {

	// Prompt the user for value
	label := "Secret value"

	prompt := promptui.Prompt{
		Label:       label,
		Default:     path,
		HideEntered: true,
	}
	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

func (sl *f) Help() {
	fmt.Printf("Prompt help\n")
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
