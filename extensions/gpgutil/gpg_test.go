package gpgutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

func buildFakeGPG(t *testing.T, dir string) string {
	t.Helper()

	source := `package main

import (
	"os"
	"strings"
)

func main() {
	path := os.Getenv("SALAINEN_TEST_GPG_ARGS")
	if path != "" {
		_ = os.WriteFile(path, []byte(strings.Join(os.Args[1:], "\n")), 0600)
	}
	_, _ = os.Stdout.WriteString("test-fingerprint\n")
}
`
	sourcePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatalf("failed to write fake gpg source: %v", err)
	}

	exe := filepath.Join(dir, "fakegpg")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	build := exec.Command("go", "build", "-o", exe, sourcePath)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build fake gpg: %v\n%s", err, output)
	}
	return exe
}

func TestRunWithPassphraseDoesNotExposePassphraseInArgs(t *testing.T) {
	dir := t.TempDir()
	exe := buildFakeGPG(t, dir)
	argsFile := filepath.Join(dir, "args.txt")
	t.Setenv("SALAINEN_TEST_GPG_ARGS", argsFile)

	const passphrase = "do-not-leak-me"
	config := Config{Executable: exe, HomeDir: filepath.Join(dir, "home")}
	if _, err := config.runWithPassphrase([]byte("input"), passphrase, "--generate-key"); err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read captured args: %v", err)
	}
	text := string(args)
	if strings.Contains(text, passphrase) {
		t.Errorf("passphrase leaked in command arguments: %q", text)
	}
	if !strings.Contains(text, "--passphrase-file") || !strings.Contains(text, "--pinentry-mode") {
		t.Errorf("passphrase-file options missing from arguments: %q", text)
	}
}

func TestValidateKeyFieldsRejectsInjection(t *testing.T) {
	if err := validateKeyFields("Name\nPassphrase: injected", "test@example.com", "0", "RSA", "2048"); err == nil {
		t.Errorf("expected newline injection to be rejected")
	}
	if err := validateKeyFields("Name", "not-an-email", "0", "RSA", "2048"); err == nil {
		t.Errorf("expected invalid email to be rejected")
	}
	if err := validateKeyFields("Name", "test@example.com", "0", "RSA", "not-a-number"); err == nil {
		t.Errorf("expected invalid key length to be rejected")
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
