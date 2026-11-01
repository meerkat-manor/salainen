package encryptedfile

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/meerkat-manor/salainen"
	"github.com/meerkat-manor/salainen/extensions/gpgutil"
	"github.com/xdg-go/pbkdf2"
	"golang.org/x/crypto/chacha20poly1305"
)

type f struct {
	RootPath      string
	Algorithm     EncryptionAlgorithm
	GPGExecutable string
	GPGHomeDir    string
	GPGRecipient  string
	GPGPassphrase string
	GPGMarker     string
	GPGName       string
	GPGEmail      string
	GPGExpireDate string
	GPGKeyType    string
	GPGKeyLength  string

	gpgPassphraseSet bool
}

// Note that if the product name changes this
// value should also change
const pemType = "SALAINEN ENCODED FILE"
const providerName = "efile"

type EncryptionAlgorithm string

const (
	ChaCha20Poly1305 EncryptionAlgorithm = "ChaCha20-Poly1305"
	GPG              EncryptionAlgorithm = "GPG"
)

func (sl *f) Init(custom interface{}) error {

	sl.Algorithm = ChaCha20Poly1305

	if custom != nil {
		if settings, ok := custom.(map[string]string); ok {

			value, exists := settings["RootPath"]
			if exists && value != "" {
				sl.RootPath = value
			}

			value, exists = settings["Algorithm"]
			if exists && value != "" {
				algo := value
				switch {
				case strings.EqualFold(algo, "ChaCha20-Poly1305"):
					sl.Algorithm = ChaCha20Poly1305
				case strings.EqualFold(algo, "GPG"), strings.EqualFold(algo, "OpenPGP"):
					sl.Algorithm = GPG
				default:
					return fmt.Errorf("encryption algorithm not recognized")
				}
			}

			value, exists = settings["GPGExecutable"]
			if exists && value != "" {
				sl.GPGExecutable = value
			}
			value, exists = settings["GPGHomeDir"]
			if exists && value != "" {
				sl.GPGHomeDir = value
			}
			value, exists = settings["GPGRecipient"]
			if exists && value != "" {
				sl.GPGRecipient = value
			}
			value, exists = settings["Recipient"]
			if exists && value != "" && sl.GPGRecipient == "" {
				sl.GPGRecipient = value
			}
			value, exists = settings["GPGPassphrase"]
			if exists {
				sl.GPGPassphrase = value
				sl.gpgPassphraseSet = true
			}
			value, exists = settings["GPGMarker"]
			if exists && value != "" {
				sl.GPGMarker = value
			}
			value, exists = settings["GPGName"]
			if exists && value != "" {
				sl.GPGName = value
			}
			value, exists = settings["GPGEmail"]
			if exists && value != "" {
				sl.GPGEmail = value
			}
			value, exists = settings["GPGExpireDate"]
			if exists && value != "" {
				sl.GPGExpireDate = value
			}
			value, exists = settings["GPGKeyType"]
			if exists && value != "" {
				sl.GPGKeyType = value
			}
			value, exists = settings["GPGKeyLength"]
			if exists && value != "" {
				sl.GPGKeyLength = value
			}
		}
	}

	sl.RootPath = strings.ReplaceAll(sl.RootPath, "{{.ProductName}}", salainen.ProductName)
	sl.GPGHomeDir = strings.ReplaceAll(sl.GPGHomeDir, "{{.ProductName}}", salainen.ProductName)

	if sl.Algorithm == GPG && sl.GPGRecipient == "" {
		marker, err := gpgutil.ReadMarker(sl.markerPath())
		if err == nil && marker.Recipient != "" {
			sl.GPGRecipient = marker.Recipient
		}
	}

	return nil
}

func (sl *f) Put(path string, value string) error {
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

	// The default encryption key (password) used
	// is the path, if not supplied
	encKey := path
	if len(parts) == 2 {
		encKey = parts[1]
	}

	switch sl.Algorithm {
	case ChaCha20Poly1305:
		return sl.Encode_ChaCha20Poly1305(encKey, fpath, value)
	case GPG:
		recipient := sl.GPGRecipient
		if len(parts) == 2 && parts[1] != "" {
			if exists, err := sl.gpgConfig(parts[1]).KeyExists(parts[1]); err == nil && exists {
				recipient = parts[1]
			}
		}
		return sl.Encode_GPG(recipient, fpath, value)
	default:
		return fmt.Errorf("algorithm (%v) not recognised", sl.Algorithm)
	}
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

	// The default encryption key (password) used
	// is the path, if not supplied
	encKey := path
	if len(parts) == 2 {
		encKey = parts[1]
	}

	i, err := os.Open(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", salainen.ErrNoSuchSecret
		}
		return "", err
	}
	defer i.Close()

	data, err := io.ReadAll(i)
	if err != nil {
		return "", err
	}

	switch sl.Algorithm {
	case ChaCha20Poly1305:
		return sl.Decode_ChaCha20Poly1305(encKey, data)
	case GPG:
		passphrase := sl.GPGPassphrase
		if len(parts) == 2 && parts[1] != "" {
			if exists, err := sl.gpgConfig(parts[1]).KeyExists(parts[1]); err != nil || !exists {
				passphrase = parts[1]
			}
		}
		return sl.Decode_GPGWithPassphrase(passphrase, data)
	default:
		return "", fmt.Errorf("algorithm (%v) not recognised", sl.Algorithm)
	}

}

func (sl *f) Encode_ChaCha20Poly1305(pass string, fpath string, value string) error {

	nonce := make([]byte, chacha20poly1305.NonceSize)
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return err
	}

	key := pbkdf2.Key([]byte(pass), nonce, 4096, chacha20poly1305.KeySize, sha256.New)

	c, err := chacha20poly1305.New(key)
	if err != nil {
		return err
	}

	data := c.Seal(nil, nonce, []byte(value), nil)

	oFile, err := os.OpenFile(fpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer oFile.Close()

	var b pem.Block
	b.Bytes = data
	b.Type = pemType
	b.Headers = map[string]string{
		"S-Nonce": base64.StdEncoding.EncodeToString(nonce),
	}

	return pem.Encode(oFile, &b)
}

func (sl *f) Decode_ChaCha20Poly1305(pass string, data []byte) (string, error) {

	blk, _ := pem.Decode(data)

	if blk == nil {
		return "", salainen.ErrInvalidSecret
	}

	if blk.Type != pemType {
		return "", salainen.ErrInvalidSecret
	}

	nonce, err := base64.StdEncoding.DecodeString(blk.Headers["S-Nonce"])
	if err != nil {
		return "", err
	}

	key := pbkdf2.Key([]byte(pass), nonce, 4096, chacha20poly1305.KeySize, sha256.New)

	c, err := chacha20poly1305.New(key)
	if err != nil {
		return "", err
	}

	raw, err := c.Open(nil, nonce, blk.Bytes, nil)
	if err != nil {
		return "", salainen.ErrInvalidSecretAccess
	}

	return string(raw), nil
}

func (sl *f) Encode_GPG(recipient string, fpath string, value string) error {

	encrypted, err := sl.gpgConfig(recipient).Encrypt([]byte(value), recipient)
	if err != nil {
		return err
	}

	parent := filepath.Dir(fpath)
	if _, err := os.Stat(parent); err != nil {
		os.MkdirAll(parent, 0700)
	}

	return os.WriteFile(fpath, encrypted, 0600)
}

func (sl *f) Decode_GPG(data []byte) (string, error) {
	return sl.Decode_GPGWithPassphrase(sl.GPGPassphrase, data)
}

func (sl *f) Decode_GPGWithPassphrase(passphrase string, data []byte) (string, error) {

	config := sl.gpgConfig("")
	config.Passphrase = passphrase

	decrypted, err := config.Decrypt(data)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func (sl *f) gpgConfig(recipient string) gpgutil.Config {
	if recipient == "" {
		recipient = sl.GPGRecipient
	}

	return gpgutil.Config{
		Executable: sl.GPGExecutable,
		HomeDir:    sl.GPGHomeDir,
		Recipient:  recipient,
		Passphrase: sl.GPGPassphrase,
	}
}

func (sl *f) markerPath() string {
	if sl.GPGMarker != "" {
		return gpgutil.ExpandPath(sl.GPGMarker)
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

	recipient := sl.GPGRecipient
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

	sl.GPGRecipient = recipient

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

	name := sl.GPGName
	if name == "" {
		name = gpgutil.Prompt("GPG key name", gpgutil.DefaultName(), false)
	}
	email := sl.GPGEmail
	if email == "" {
		email = gpgutil.Prompt("GPG key email", gpgutil.DefaultEmail(name), false)
	}
	passphrase := sl.GPGPassphrase
	if !sl.gpgPassphraseSet {
		passphrase = gpgutil.Prompt("GPG key passphrase (blank for none)", "", true)
	}
	expireDate := sl.GPGExpireDate
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
		KeyType:    sl.GPGKeyType,
		KeyLength:  sl.GPGKeyLength,
	})
	if err != nil {
		return err
	}

	sl.GPGRecipient = fingerprint
	sl.GPGPassphrase = passphrase
	sl.gpgPassphraseSet = true

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
	fmt.Printf("Encrypted file help\n\n")
	fmt.Printf("An encrypted file can be used as a secret provider by\n")
	fmt.Printf("using the prefix 'efile:' followed by the\n")
	fmt.Printf("file name in the configured directory.  The contents in\n")
	fmt.Printf("the file is the encrypted secret which is decoded.\n")
	fmt.Printf("\n")
	fmt.Printf("The security of the secret is only as good as the access\n")
	fmt.Printf("granted to the file and the encoding key.\n")
	fmt.Printf("\n")
	fmt.Printf("The supported algorithms are ChaCha20-Poly1305 (default) and GPG.\n")
	fmt.Printf("For GPG the public key is used to encrypt the secret and the\n")
	fmt.Printf("matching private key is used to decrypt it.  The recipient can\n")
	fmt.Printf("be set with GPGRecipient or supplied as 'efile:file|recipient'.\n")
	fmt.Printf("\n")
	fmt.Printf("Use the option '-provider -init efile' to create the provider\n")
	fmt.Printf("directory and, when using GPG, generate a key pair.\n")
	fmt.Printf("\n")
	fmt.Printf("For more information please see %s/extensions/encryptedfile/ \n", salainen.SourceForgeURL)
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
