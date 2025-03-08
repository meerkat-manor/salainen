package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/meerkat-manor/salainen"
)

type f struct {
	RootPath string
}

const providerName = "file"

func (sl *f) Init(custom interface{}) error {

	if custom != nil {
		if settings, ok := custom.(map[string]string); ok {

			value, exists := settings["RootPath"]
			if exists && value != "" {
				sl.RootPath = value
			}
		}
	}

	sl.RootPath = strings.ReplaceAll(sl.RootPath, "{{.ProductName}}", salainen.ProductName)

	return nil
}

func (sl *f) Put(path string, val string) error {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	if sl.RootPath != "" {
		fpath = filepath.Join(sl.RootPath, fpath)
	}
	if strings.HasPrefix(fpath, "~/") || strings.HasPrefix(fpath, "~\\") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		fpath = filepath.Join(homeDir, fpath[2:])
	}

	parent := filepath.Dir(fpath)
	if _, err := os.Stat(parent); err != nil {
		os.MkdirAll(parent, os.ModeDir)
	}

	err := os.WriteFile(fpath, []byte(val), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (sl *f) Get(path string) (string, error) {
	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	if sl.RootPath != "" {
		fpath = filepath.Join(sl.RootPath, fpath)
	}
	if strings.HasPrefix(fpath, "~/") || strings.HasPrefix(fpath, "~\\") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		fpath = filepath.Join(homeDir, fpath[2:])
	}

	parent := filepath.Dir(fpath)
	if _, err := os.Stat(parent); err != nil {
		return "", fmt.Errorf("directory '%s' does not exist", parent)
	}

	data, err := os.ReadFile(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", salainen.ErrNoSuchSecret
		}
		return "", err
	}

	// Strip last character EOL/EOF
	return string(data), nil
}

func (sl *f) Help() {
	fmt.Printf("File help\n\n")
	fmt.Printf("A file can be used as a secret provider by using\n")
	fmt.Printf("the prefix 'file:' followed by the file name\n")
	fmt.Printf("in the configured directory.  The contents in\n")
	fmt.Printf("the file is the secret.\n")
	fmt.Printf("\n")
	fmt.Printf("The security of the secret is only as good as the access\n")
	fmt.Printf("granted to the file.\n")
	fmt.Printf("\n")
	fmt.Printf("The special value of '~' is recognised as your home directory.\n")
	fmt.Printf("You can store secrets in your home directory by specifying a\n")
	fmt.Printf("configured 'RootPath' such as '~/.secrets/%s'\n", "salainen") // TODO improve
	fmt.Printf("\n")
	fmt.Printf("For more information please see %s/extensions/file/ \n", salainen.SourceForgeURL)
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
