package encryptedfile

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/meerkat-manor/salainen"
	"github.com/meerkat-manor/salainen/extensions/gpgutil"
)

func TestGPGEncFilePutGet(t *testing.T) {

	if _, err := exec.LookPath(gpgutil.DefaultExecutable); err != nil {
		t.Skip("gpg executable not available")
	}

	gpgHome := t.TempDir()
	rootPath := t.TempDir()

	gpgConfig := gpgutil.Config{HomeDir: gpgHome}
	fingerprint, err := gpgConfig.GenerateKey(gpgutil.KeyRequest{
		Name:      "Salainen EFile Test",
		Email:     "salainen-efile-test@example.com",
		KeyType:   "RSA",
		KeyLength: "2048",
	})
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	custom := map[string]string{
		"RootPath":      rootPath,
		"Algorithm":     "GPG",
		"GPGHomeDir":    gpgHome,
		"GPGRecipient":  fingerprint,
		"GPGMarker":     filepath.Join(rootPath, "marker.json"),
		"GPGPassphrase": "",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	initializer, ok := sto.(salainen.SecretStorageInitializer)
	if !ok {
		t.Fatal("storage does not support initialisation")
	}
	initialised, err := initializer.Initialised()
	if err != nil {
		t.Fatalf("failed to check initialisation: %v", err)
	}
	if !initialised {
		t.Error("expected provider to be initialised")
	}

	key := "testcaseGpgPutGet.txt"
	secret := "mysecretgpg01"
	if err := sto.Put(key, secret); err != nil {
		t.Fatalf("failed to put value: %v", err)
	}

	value, err := sto.Get(key)
	if err != nil || value != secret {
		t.Errorf("failed to get value: '%s' %v", value, err)
	}

	keyWithPassphrase := "testcaseGpgPutGetPass.txt|anything"
	if err := sto.Put(keyWithPassphrase, secret); err != nil {
		t.Fatalf("failed to put value with path passphrase: %v", err)
	}

	value, err = sto.Get(keyWithPassphrase)
	if err != nil || value != secret {
		t.Errorf("failed to get value with path passphrase: '%s' %v", value, err)
	}
}
