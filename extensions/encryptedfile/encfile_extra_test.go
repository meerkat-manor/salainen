package encryptedfile

import (
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meerkat-manor/salainen"
)

func newTempEncStorage(t *testing.T, custom map[string]string) salainen.SecretStorage {
	t.Helper()
	settings := map[string]string{"RootPath": t.TempDir()}
	for key, value := range custom {
		settings[key] = value
	}
	storage, err := New("", settings)
	if err != nil {
		t.Fatalf("failed to create encrypted file storage: %v", err)
	}
	return storage
}

func readEncryptedFile(t *testing.T, root, key string) ([]byte, *pem.Block) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, key))
	if err != nil {
		t.Fatalf("failed to read encrypted file: %v", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatalf("encrypted file is not valid PEM")
	}
	return data, block
}

func TestEncFileRoundTripInTempDir(t *testing.T) {

	storage := newTempEncStorage(t, nil)

	if err := storage.Put("secret", "plain-value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}
	value, err := storage.Get("secret")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if value != "plain-value" {
		t.Errorf("unexpected value: %q", value)
	}
}

func TestEncFileMissingSecretSentinel(t *testing.T) {

	storage := newTempEncStorage(t, nil)

	if _, err := storage.Get("does-not-exist"); !errors.Is(err, salainen.ErrNoSuchSecret) {
		t.Errorf("expected ErrNoSuchSecret, got %v", err)
	}
}

func TestEncFileCiphertextFormat(t *testing.T) {

	root := t.TempDir()
	storage, err := New("", map[string]string{"RootPath": root})
	if err != nil {
		t.Fatalf("failed to create encrypted file storage: %v", err)
	}

	const secret = "value-that-must-not-appear-in-cleartext"
	if err := storage.Put("secret", secret); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	data, block := readEncryptedFile(t, root, "secret")
	if block.Type != pemType {
		t.Errorf("unexpected PEM block type: %q", block.Type)
	}
	if strings.Contains(string(data), secret) {
		t.Errorf("plaintext found in encrypted file")
	}

	nonce, err := base64.StdEncoding.DecodeString(block.Headers["S-Nonce"])
	if err != nil {
		t.Fatalf("invalid S-Nonce header: %v", err)
	}
	if len(nonce) != 12 {
		t.Errorf("expected a 12-byte nonce, got %d bytes", len(nonce))
	}
}

func TestEncFileStrongKDFHeaders(t *testing.T) {
	root := t.TempDir()
	storage, err := New("", map[string]string{"RootPath": root})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	if err := storage.Put("secret", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	_, block := readEncryptedFile(t, root, "secret")
	if block.Headers["S-Version"] != "2" {
		t.Errorf("unexpected format version %q", block.Headers["S-Version"])
	}
	if block.Headers["S-KDF"] != keyDerivation {
		t.Errorf("unexpected KDF %q", block.Headers["S-KDF"])
	}
	salt, err := base64.StdEncoding.DecodeString(block.Headers["S-Salt"])
	if err != nil {
		t.Fatalf("invalid S-Salt header: %v", err)
	}
	if len(salt) != 16 {
		t.Errorf("expected a 16-byte salt, got %d bytes", len(salt))
	}
}

func TestEncFileConfiguredPassphrase(t *testing.T) {
	root := t.TempDir()
	storage, err := New("", map[string]string{
		"RootPath":   root,
		"Passphrase": "configured-passphrase",
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	if err := storage.Put("secret", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}
	value, err := storage.Get("secret")
	if err != nil || value != "value" {
		t.Errorf("configured passphrase round trip failed: %q %v", value, err)
	}

	if err := storage.Put("override|explicit-passphrase", "value"); err != nil {
		t.Fatalf("unexpected override put error: %v", err)
	}
	value, err = storage.Get("override|explicit-passphrase")
	if err != nil || value != "value" {
		t.Errorf("explicit passphrase round trip failed: %q %v", value, err)
	}
}

func TestEncFileDetectsTampering(t *testing.T) {

	root := t.TempDir()
	storage, err := New("", map[string]string{"RootPath": root})
	if err != nil {
		t.Fatalf("failed to create encrypted file storage: %v", err)
	}

	if err := storage.Put("secret", "original"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	_, block := readEncryptedFile(t, root, "secret")
	if len(block.Bytes) == 0 {
		t.Fatalf("empty ciphertext")
	}
	block.Bytes[0] ^= 0xFF

	tampered := pem.EncodeToMemory(block)
	if err := os.WriteFile(filepath.Join(root, "secret"), tampered, 0o600); err != nil {
		t.Fatalf("failed to write tampered file: %v", err)
	}

	if _, err := storage.Get("secret"); !errors.Is(err, salainen.ErrInvalidSecretAccess) {
		t.Errorf("expected ErrInvalidSecretAccess, got %v", err)
	}
}

func TestEncFileInvalidPEM(t *testing.T) {

	root := t.TempDir()
	storage, err := New("", map[string]string{"RootPath": root})
	if err != nil {
		t.Fatalf("failed to create encrypted file storage: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "secret"), []byte("not a pem file"), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if _, err := storage.Get("secret"); !errors.Is(err, salainen.ErrInvalidSecret) {
		t.Errorf("expected ErrInvalidSecret, got %v", err)
	}
}

func TestEncFileUsesUniqueNonces(t *testing.T) {

	root := t.TempDir()
	storage, err := New("", map[string]string{"RootPath": root})
	if err != nil {
		t.Fatalf("failed to create encrypted file storage: %v", err)
	}

	if err := storage.Put("first", "same-value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}
	if err := storage.Put("second", "same-value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	_, first := readEncryptedFile(t, root, "first")
	_, second := readEncryptedFile(t, root, "second")

	if first.Headers["S-Nonce"] == second.Headers["S-Nonce"] {
		t.Errorf("two encryptions reused the same nonce")
	}
}

func TestEncFileWrongPassphrase(t *testing.T) {

	storage := newTempEncStorage(t, nil)

	if err := storage.Put("secret|correct-passphrase", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	if _, err := storage.Get("secret|wrong-passphrase"); !errors.Is(err, salainen.ErrInvalidSecretAccess) {
		t.Errorf("expected ErrInvalidSecretAccess, got %v", err)
	}
}

func TestEncFileProductNameSubstitution(t *testing.T) {

	root := t.TempDir()
	storage, err := New("", map[string]string{
		"RootPath": filepath.Join(root, "{{.ProductName}}"),
	})
	if err != nil {
		t.Fatalf("failed to create encrypted file storage: %v", err)
	}

	if err := storage.Put("secret", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, salainen.ProductName, "secret")); err != nil {
		t.Errorf("expected substituted directory to be used: %v", err)
	}
}

func TestEncFileAlgorithmAliases(t *testing.T) {

	for _, algorithm := range []string{"ChaCha20-Poly1305", "chacha20-poly1305"} {
		t.Run(algorithm, func(t *testing.T) {
			storage, err := New("", map[string]string{
				"RootPath":  t.TempDir(),
				"Algorithm": algorithm,
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if storage == nil {
				t.Errorf("expected a storage instance")
			}
		})
	}
}
