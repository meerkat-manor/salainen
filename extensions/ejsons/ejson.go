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
	"github.com/meerkat-manor/salainen/extensions/gpgutil"
)

const (
	DefaultElement = "password"
)

type EncryptionAlgorithm string

const (
	EJSON EncryptionAlgorithm = "EJSON"
	GPG   EncryptionAlgorithm = "GPG"
)

type gpg struct {
	Executable string
	HomeDir    string
	Recipient  string
	Passphrase string
	Marker     string
	Name       string
	Email      string
	ExpireDate string
	KeyType    string
	KeyLength  string

	passphraseSet bool
}

type f struct {
	RootPath    string
	KeyDir      string
	PrivateKey  string
	PublicKey   string
	ElementName string
	Algorithm   EncryptionAlgorithm
	GPG         gpg
}

const providerName = "ejson"

func (sl *f) Init(custom interface{}) error {

	sl.Algorithm = EJSON

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
			value, exists = settings["Algorithm"]
			if exists && value != "" {
				switch {
				case strings.EqualFold(value, "EJSON"), strings.EqualFold(value, "native"):
					sl.Algorithm = EJSON
				case strings.EqualFold(value, "GPG"), strings.EqualFold(value, "OpenPGP"):
					sl.Algorithm = GPG
				default:
					return fmt.Errorf("encryption algorithm not recognized")
				}
			}
			value, exists = settings["GPGExecutable"]
			if exists && value != "" {
				sl.GPG.Executable = value
			}
			value, exists = settings["GPGHomeDir"]
			if exists && value != "" {
				sl.GPG.HomeDir = value
			}
			value, exists = settings["GPGRecipient"]
			if exists && value != "" {
				sl.GPG.Recipient = value
			}
			value, exists = settings["Recipient"]
			if exists && value != "" && sl.GPG.Recipient == "" {
				sl.GPG.Recipient = value
			}
			value, exists = settings["GPGPassphrase"]
			if exists {
				sl.GPG.Passphrase = value
				sl.GPG.passphraseSet = true
			}
			value, exists = settings["GPGMarker"]
			if exists && value != "" {
				sl.GPG.Marker = value
			}
			value, exists = settings["GPGName"]
			if exists && value != "" {
				sl.GPG.Name = value
			}
			value, exists = settings["GPGEmail"]
			if exists && value != "" {
				sl.GPG.Email = value
			}
			value, exists = settings["GPGExpireDate"]
			if exists && value != "" {
				sl.GPG.ExpireDate = value
			}
			value, exists = settings["GPGKeyType"]
			if exists && value != "" {
				sl.GPG.KeyType = value
			}
			value, exists = settings["GPGKeyLength"]
			if exists && value != "" {
				sl.GPG.KeyLength = value
			}

		}
	}

	if sl.ElementName == "" {
		sl.ElementName = DefaultElement
	}

	if sl.Algorithm != GPG && sl.PrivateKey != "" {
		if strings.HasPrefix(sl.PrivateKey, (providerName + ":")) {
			return fmt.Errorf("error fetching %s private key with looping detected", providerName)
		}

		privateKey, errS := salainen.Get(sl.PrivateKey)
		if errS != nil || privateKey == "" {
			return fmt.Errorf("error fetching %s private key.  More information: %v", providerName, errS)
		}

		sl.PrivateKey = privateKey
	}

	sl.RootPath = strings.ReplaceAll(sl.RootPath, "{{.ProductName}}", salainen.ProductName)
	sl.GPG.HomeDir = strings.ReplaceAll(sl.GPG.HomeDir, "{{.ProductName}}", salainen.ProductName)

	if sl.Algorithm == GPG && sl.GPG.Recipient == "" {
		marker, err := gpgutil.ReadMarker(sl.markerPath())
		if err == nil && marker.Recipient != "" {
			sl.GPG.Recipient = marker.Recipient
		}
	}

	return nil
}

func (sl *f) Put(path string, val string) error {

	if sl.Algorithm == EJSON && sl.PublicKey == "" {
		return fmt.Errorf("no public encryption key provided")
	}

	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	recipient := sl.GPG.Recipient
	passphrase := sl.GPG.Passphrase
	if sl.Algorithm == GPG && len(parts) == 2 && parts[1] != "" {
		if exists, err := sl.gpgConfig(parts[1]).KeyExists(parts[1]); err == nil && exists {
			recipient = parts[1]
		} else {
			passphrase = parts[1]
		}
	}
	if sl.Algorithm == GPG && recipient == "" {
		return fmt.Errorf("no GPG recipient configured. Run the provider -init option or set GPGRecipient")
	}

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
			if sl.Algorithm == GPG {
				config := sl.gpgConfig(recipient)
				config.Passphrase = passphrase
				contents, err = config.Decrypt(contents)
				if err != nil {
					return fmt.Errorf("could not decrypt existing JSON file. Error: %v", err)
				}
			}
		}
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(contents, &jsonData)
	if err != nil {
		return fmt.Errorf("could not parse existing JSON file. Error: %v", err)
	}
	if jsonData == nil {
		jsonData = map[string]interface{}{}
	}

	if sl.Algorithm == EJSON {
		// Check elements and add as necessary
		chkVal, exists := jsonData["_public_key"]
		if exists {
			if chkVal == "" {
				jsonData["_public_key"] = sl.PublicKey
			}
		} else {
			jsonData["_public_key"] = sl.PublicKey
		}
	}

	jsonData[sl.ElementName] = val

	contents, err = json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("could not marshal JSON. Error: %v", err)
	}

	if sl.Algorithm == GPG {
		contents, err = sl.gpgConfig(recipient).Encrypt(contents, recipient)
		if err != nil {
			return fmt.Errorf("could not encrypt JSON. Error: %v", err)
		}
	} else {
		rdr := bytes.NewReader(contents)
		var outBuffer bytes.Buffer

		count, err := ejson.Encrypt(rdr, &outBuffer)
		if err != nil || count < 1 {
			return fmt.Errorf("could not encrypt JSON. Error: %v", err)
		}
		contents = outBuffer.Bytes()
	}

	fileMode := os.FileMode(0600)
	if stat, err := os.Stat(fpath); err == nil {
		fileMode = stat.Mode()
	}

	parent := filepath.Dir(fpath)
	if _, err := os.Stat(parent); err != nil {
		os.MkdirAll(parent, 0700)
	}

	if err := os.WriteFile(fpath, contents, fileMode); err != nil {
		return fmt.Errorf("could not write encrypt JSON file. Error: %v", err)
	}

	return nil
}

func (sl *f) Get(path string) (string, error) {

	if sl.Algorithm != GPG && sl.KeyDir == "" && sl.PrivateKey == "" {
		return "", fmt.Errorf("no decryption (KeyDir or PrivateKey) configuration values provided")
	}

	parts := strings.SplitN(path, "|", 2)
	fpath := parts[0]

	passphrase := sl.GPG.Passphrase
	if sl.Algorithm == GPG && len(parts) == 2 && parts[1] != "" {
		if exists, err := sl.gpgConfig(parts[1]).KeyExists(parts[1]); err != nil || !exists {
			passphrase = parts[1]
		}
	}

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
	var data []byte
	var err error

	if sl.Algorithm == GPG {
		contents, err := os.ReadFile(fpath)
		if err != nil {
			return "", err
		}
		config := sl.gpgConfig("")
		config.Passphrase = passphrase
		data, err = config.Decrypt(contents)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt file contents. %v", err)
		}
	} else {
		data, err = ejson.DecryptFile(fpath, sl.KeyDir, sl.PrivateKey)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt file contents. %v", err)
		}
	}

	// Get required element
	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return "", fmt.Errorf("could not parse existing JSON. Error: %v", err)
	}

	value, exists := jsonData[sl.ElementName]
	if exists {
		if text, ok := value.(string); ok {
			return text, nil
		}
		return fmt.Sprintf("%v", value), nil
	} else {
		return "", nil
	}

}

func (sl *f) gpgConfig(recipient string) gpgutil.Config {
	if recipient == "" {
		recipient = sl.GPG.Recipient
	}

	return gpgutil.Config{
		Executable: sl.GPG.Executable,
		HomeDir:    sl.GPG.HomeDir,
		Recipient:  recipient,
		Passphrase: sl.GPG.Passphrase,
	}
}

func (sl *f) markerPath() string {
	if sl.GPG.Marker != "" {
		return gpgutil.ExpandPath(sl.GPG.Marker)
	}
	return gpgutil.MarkerPath(sl.RootPath, providerName)
}

func (sl *f) rootPath() string {
	if sl.RootPath == "" {
		return ""
	}
	return gpgutil.ExpandPath(sl.RootPath)
}

func (sl *f) Initialised() (bool, error) {

	if sl.Algorithm != GPG {
		if sl.RootPath == "" {
			return true, nil
		}
		_, err := os.Stat(sl.rootPath())
		if err == nil {
			return true, nil
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	recipient := sl.GPG.Recipient
	if recipient == "" {
		marker, err := gpgutil.ReadMarker(sl.markerPath())
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		recipient = marker.Recipient
	}

	if recipient == "" {
		return false, nil
	}

	exists, err := sl.gpgConfig(recipient).KeyExists(recipient)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	sl.GPG.Recipient = recipient

	return true, nil
}

func (sl *f) Initialise() error {

	if sl.Algorithm != GPG {
		if sl.RootPath != "" {
			return os.MkdirAll(sl.rootPath(), 0700)
		}
		return nil
	}

	initialised, err := sl.Initialised()
	if err != nil {
		return err
	}
	if initialised {
		return nil
	}

	name := sl.GPG.Name
	if name == "" {
		name = gpgutil.Prompt("GPG key name", gpgutil.DefaultName(), false)
	}
	email := sl.GPG.Email
	if email == "" {
		email = gpgutil.Prompt("GPG key email", gpgutil.DefaultEmail(name), false)
	}
	passphrase := sl.GPG.Passphrase
	if !sl.GPG.passphraseSet {
		passphrase = gpgutil.Prompt("GPG key passphrase (blank for none)", "", true)
	}
	expireDate := sl.GPG.ExpireDate
	if expireDate == "" {
		expireDate = gpgutil.Prompt("GPG key expiry (0 = never)", "0", false)
	}

	gpgConfig := sl.gpgConfig("")
	gpgConfig.Passphrase = passphrase

	fingerprint, err := gpgConfig.GenerateKey(gpgutil.KeyRequest{
		Name:       name,
		Email:      email,
		Passphrase: passphrase,
		ExpireDate: expireDate,
		KeyType:    sl.GPG.KeyType,
		KeyLength:  sl.GPG.KeyLength,
	})
	if err != nil {
		return err
	}

	sl.GPG.Recipient = fingerprint
	sl.GPG.Passphrase = passphrase
	sl.GPG.passphraseSet = true

	marker := gpgutil.NewMarker(name, email, fingerprint)
	if err := gpgutil.WriteMarker(sl.markerPath(), marker); err != nil {
		return err
	}

	fmt.Printf("GPG key generated for provider '%s'\n", providerName)
	fmt.Printf("Recipient: %s\n", fingerprint)
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Email: %s\n", email)

	return nil
}

func (sl *f) Initialized() (bool, error) {
	return sl.Initialised()
}

func (sl *f) Initialize() error {
	return sl.Initialise()
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
	fmt.Printf("The supported algorithms are EJSON (default) and GPG.\n")
	fmt.Printf("When GPG is selected the complete JSON file is encrypted with\n")
	fmt.Printf("the configured GPG recipient and decrypted with the matching\n")
	fmt.Printf("private key.  The recipient can be set with GPGRecipient or\n")
	fmt.Printf("supplied as 'ejson:file|recipient'.\n")
	fmt.Printf("\n")
	fmt.Printf("Use the option '-provider -init ejson' to create the provider\n")
	fmt.Printf("directory and, when using GPG, generate a key pair.\n")
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
