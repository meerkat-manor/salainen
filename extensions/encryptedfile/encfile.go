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
	"github.com/xdg-go/pbkdf2"
	"golang.org/x/crypto/chacha20poly1305"
)

type f struct {
	RootPath  string
	Algorithm EncryptionAlgorithm
}

const pemType = "SALAINEN ENCODED FILE"
const providerName = "efile"

type EncryptionAlgorithm string

const (
	ChaCha20Poly1305 EncryptionAlgorithm = "ChaCha20-Poly1305"
)

func (sl *f) Init(custom interface{}) error {

	sl.Algorithm = ChaCha20Poly1305

	if custom != nil {
		settings := custom.(map[string]interface{})

		value, exists := settings["RootPath"]
		if exists && value.(string) != "" {
			sl.RootPath = value.(string)
		}

		value, exists = settings["Algorithm"]
		if exists && value.(string) != "" {
			algo := value.(string)
			switch algo {
			case "ChaCha20-Poly1305":
				sl.Algorithm = ChaCha20Poly1305
			default:
				return fmt.Errorf("encryption algorithm not recognized")
			}
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

	// The default encryption key (password) used
	// is the path, if not supplied
	encKey := path
	if len(parts) == 2 {
		encKey = parts[1]
	}

	switch sl.Algorithm {
	case ChaCha20Poly1305:
		return sl.Encode_ChaCha20Poly1305(encKey, fpath, value)
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

	data, err := io.ReadAll(i)
	if err != nil {
		return "", err
	}

	switch sl.Algorithm {
	case ChaCha20Poly1305:
		return sl.Decode_ChaCha20Poly1305(encKey, data)
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
	fmt.Printf("For more information please see %s/extensions/encryptedfile/README.md \n", salainen.SourceForgeURL)
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
