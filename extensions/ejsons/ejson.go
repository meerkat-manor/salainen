package ejsons

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shopify/ejson"
	"github.com/meerkat-manor/salainen"
)

const (
	DefaultElement = "password"
)

type f struct {
	RootPath    string
	KeyDir      string
	PrivateKey  string
	PublicKey   string
	ElementName string
}

const providerName = "ejson"

func (sl *f) Init(custom interface{}) error {

	if custom != nil {
		if settings, ok := custom.(map[string]string); ok {

			value, exists := settings["RootPath"]
			if exists && value != "" {
				sl.RootPath = value
			}
			value, exists = settings["KeyDir"]
			if exists && value != "" {
				sl.KeyDir = value
			}
			value, exists = settings["PrivateKey"]
			if exists && value != "" {
				sl.PrivateKey = value
			}
			value, exists = settings["PublicKey"]
			if exists && value != "" {
				sl.PublicKey = value
			}
			value, exists = settings["ElementName"]
			if exists && value != "" {
				sl.ElementName = value
			}

		}
	}

	if sl.ElementName == "" {
		sl.ElementName = DefaultElement
	}

	if sl.PrivateKey != "" {
		if strings.HasPrefix(sl.PrivateKey, (providerName + ":")) {
			return fmt.Errorf("error fetching %s private key with looping detected", providerName)
		}

		privateKey, errS := salainen.Get(sl.PrivateKey)
		if errS != nil || privateKey == "" {
			return fmt.Errorf("error fetching %s private key.  More information: %v", providerName, errS)
		}

		sl.PrivateKey = privateKey
	}

	return nil
}

func (sl *f) Put(path string, val string) error {

	if sl.PublicKey == "" {
		return fmt.Errorf("no public encryption key provided")
	}

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

	var contents []byte
	var err error

	if fi, err := os.Stat(fpath); err != nil {
		contents = []byte("{}")
	} else {
		if fi.IsDir() {
			return fmt.Errorf("supplied file name is a directory")
		}

		if fi.Size() < 1 {
			contents = []byte("{}")
		} else {
			contents, err = os.ReadFile(fpath)
			if err != nil {
				return fmt.Errorf("could not read existing JSON file. Error: %v", err)
			}
		}
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(contents, &jsonData)
	if err != nil {
		return fmt.Errorf("could not parse existing JSON file. Error: %v", err)
	}

	// Check elements and add as necessary
	chkVal, exists := jsonData["_public_key"]
	if exists {
		if chkVal == "" {
			jsonData["_public_key"] = sl.PublicKey
		}
	} else {
		jsonData["_public_key"] = sl.PublicKey
	}

	jsonData[sl.ElementName] = val

	contents, err = json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("could not marshal JSON. Error: %v", err)
	}

	rdr := bytes.NewReader(contents)
	var outBuffer bytes.Buffer

	count, err := ejson.Encrypt(rdr, &outBuffer)
	if err != nil || count < 1 {
		return fmt.Errorf("could not encrypt JSON. Error: %v", err)
	}

	var fileMode os.FileMode
	if stat, err := os.Stat(fpath); err == nil {
		fileMode = stat.Mode()
	} else {
		fileMode = os.ModePerm
	}

	parent := filepath.Dir(fpath)
	if _, err := os.Stat(parent); err != nil {
		os.MkdirAll(parent, os.ModeDir)
	}

	if err := os.WriteFile(fpath, outBuffer.Bytes(), fileMode); err != nil {
		return fmt.Errorf("could not write encrypt JSON file. Error: %v", err)
	}

	return nil
}

func (sl *f) Get(path string) (string, error) {

	if sl.KeyDir == "" && sl.PrivateKey == "" {
		return "", fmt.Errorf("no decryption (KeyDir or PrivateKey) configuration values provided")
	}

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

	if _, err := os.Stat(fpath); err != nil {
		if os.IsNotExist(err) {
			return "", salainen.ErrNoSuchSecret
		}
		return "", err
	}

	// Decrypt file
	data, err := ejson.DecryptFile(fpath, sl.KeyDir, sl.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt file contents. %v", err)
	}

	// Get required element
	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return "", fmt.Errorf("could not parse existing JSON. Error: %v", err)
	}

	value, exists := jsonData[sl.ElementName]
	if exists {
		return value.(string), nil
	} else {
		return "", nil
	}

}

func (sl *f) Help() {
	fmt.Printf("ejson help\n\n")
	fmt.Printf("A JSON file can be used as a secret provider by using\n")
	fmt.Printf("the prefix 'ejson:' followed by the file name\n")
	fmt.Printf("in the configured directory.  The contents in\n")
	fmt.Printf("the ejson file are the encoded secret.\n")
	fmt.Printf("\n")
	fmt.Printf("The ejson file is a JSON format except that values (not names)\n")
	fmt.Printf("are encrypted when stored in the file, and you need to \n")
	fmt.Printf("decrypt before use.\n")
	fmt.Printf("\n")
	fmt.Printf("For more information please see %s/extensions/ejsons/ \n", salainen.SourceForgeURL)
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
