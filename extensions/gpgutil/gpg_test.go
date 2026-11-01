package gpgutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func testConfig(t *testing.T) Config {
	t.Helper()

	if _, err := exec.LookPath(DefaultExecutable); err != nil {
		t.Skip("gpg executable not available")
	}

	home := t.TempDir()
	return Config{
		HomeDir:    home,
		Passphrase: "salainen-test-passphrase",
	}
}

func TestGenerateEncryptDecrypt(t *testing.T) {

	config := testConfig(t)

	fingerprint, err := config.GenerateKey(KeyRequest{
		Name:      "Salainen Test",
		Email:     "salainen-test@example.com",
		KeyType:   "RSA",
		KeyLength: "2048",
	})
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	if fingerprint == "" {
		t.Fatal("empty fingerprint")
	}

	exists, err := config.KeyExists(fingerprint)
	if err != nil {
		t.Fatalf("failed to check key: %v", err)
	}
	if !exists {
		t.Fatal("generated key not found")
	}

	encrypted, err := config.Encrypt([]byte("my secret"), fingerprint)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	decrypted, err := config.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}
	if string(decrypted) != "my secret" {
		t.Errorf("unexpected decrypted value: %s", string(decrypted))
	}
}

func TestMarker(t *testing.T) {

	root := t.TempDir()
	markerPath := MarkerPath(root, "efile")

	if _, err := ReadMarker(markerPath); !os.IsNotExist(err) {
		t.Fatalf("expected missing marker, got: %v", err)
	}

	marker := NewMarker("Name", "email@example.com", "fingerprint")
	if err := WriteMarker(markerPath, marker); err != nil {
		t.Fatalf("failed to write marker: %v", err)
	}

	read, err := ReadMarker(markerPath)
	if err != nil {
		t.Fatalf("failed to read marker: %v", err)
	}
	if read.Fingerprint != marker.Fingerprint || read.Email != marker.Email {
		t.Errorf("marker mismatch: %+v", read)
	}

	if filepath.Base(markerPath) != ".salainen-efile-gpg.json" {
		t.Errorf("unexpected marker path: %s", markerPath)
	}
}
