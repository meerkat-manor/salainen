package gpgutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/meerkat-manor/salainen"
)

const DefaultExecutable = "gpg"

var (
	namePattern   = regexp.MustCompile(`^[\p{L}\p{N} .,'_-]{1,100}$`)
	emailPattern  = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+$`)
	expiryPattern = regexp.MustCompile(`^(0|[1-9][0-9]*[dwmy]?)$`)
)

type Config struct {
	Executable string
	HomeDir    string
	Recipient  string
	Passphrase string
}

type KeyRequest struct {
	Name       string
	Email      string
	Passphrase string
	ExpireDate string
	KeyType    string
	KeyLength  string
}

type Marker struct {
	Recipient   string `json:"recipient"`
	Fingerprint string `json:"fingerprint"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Algorithm   string `json:"algorithm"`
	Created     string `json:"created"`
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

func MarkerPath(rootPath string, providerName string) string {
	root := ExpandPath(rootPath)
	if root == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			root = filepath.Join(homeDir, ".secrets", salainen.ProductName)
		}
	}
	return filepath.Join(root, "."+salainen.ProductName+"-"+providerName+"-gpg.json")
}

func ReadMarker(path string) (*Marker, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	marker := Marker{}
	if err := json.Unmarshal(data, &marker); err != nil {
		return nil, fmt.Errorf("could not parse GPG marker file '%s'. Error: %v", path, err)
	}

	return &marker, nil
}

func WriteMarker(path string, marker Marker) error {
	parent := filepath.Dir(path)
	if _, err := os.Stat(parent); err != nil {
		if err := os.MkdirAll(parent, 0o700); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func Prompt(label string, defaultValue string, secret bool) string {
	result, err := PromptResult(label, defaultValue, secret)
	if err != nil {
		return defaultValue
	}
	return result
}

func PromptResult(label string, defaultValue string, secret bool) (string, error) {
	prompt := promptui.Prompt{
		Label:       label,
		Default:     defaultValue,
		HideEntered: secret,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	if result == "" {
		return defaultValue, nil
	}
	return result, nil
}

func ConfirmUnprotectedKey() (bool, error) {
	answer, err := PromptResult("Create an unprotected GPG key? Type 'yes' to confirm", "", false)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(answer), "yes"), nil
}

func DefaultName() string {
	if value := os.Getenv("USER"); value != "" {
		return value
	}
	if value := os.Getenv("USERNAME"); value != "" {
		return value
	}
	return salainen.ProductName
}

func DefaultEmail(name string) string {
	if name == "" {
		name = DefaultName()
	}
	return name + "@localhost"
}

func (c Config) execName() string {
	if c.Executable != "" {
		return c.Executable
	}
	return DefaultExecutable
}

func (c Config) baseArgs() []string {
	args := []string{"--batch", "--yes"}
	if c.HomeDir != "" {
		args = append(args, "--homedir", ExpandPath(c.HomeDir))
	}
	return args
}

func (c Config) checkExecutable() error {
	if c.Executable != "" && !filepath.IsAbs(c.Executable) {
		return fmt.Errorf("GPG executable must be an absolute path when configured")
	}
	if _, err := exec.LookPath(c.execName()); err != nil {
		return fmt.Errorf("GPG executable '%s' not found. Please install GPG and ensure it is on the PATH", c.execName())
	}
	if c.HomeDir != "" {
		if err := os.MkdirAll(ExpandPath(c.HomeDir), 0o700); err != nil {
			return fmt.Errorf("could not create GPG home directory '%s'. Error: %v", c.HomeDir, err)
		}
	}
	return nil
}

func (c Config) run(input []byte, args ...string) ([]byte, error) {
	return c.runWithPassphrase(input, "", args...)
}

func (c Config) runWithPassphrase(input []byte, passphrase string, args ...string) ([]byte, error) {
	if err := c.checkExecutable(); err != nil {
		return nil, err
	}

	allArgs := make([]string, 0, len(args)+4)
	if passphrase != "" {
		file, err := os.CreateTemp("", "salainen-gpg-passphrase-*")
		if err != nil {
			return nil, err
		}
		passphrasePath := file.Name()
		defer os.Remove(passphrasePath)
		if err := file.Chmod(0o600); err != nil {
			file.Close()
			return nil, err
		}
		if _, err := file.WriteString(passphrase + "\n"); err != nil {
			file.Close()
			return nil, err
		}
		if err := file.Close(); err != nil {
			return nil, err
		}
		allArgs = append(allArgs, "--pinentry-mode", "loopback", "--passphrase-file", passphrasePath)
	}
	allArgs = append(allArgs, args...)

	cmd := exec.Command(c.execName(), allArgs...)
	cmd.Stdin = bytes.NewReader(input)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("GPG command failed. Error: %v. Details: %s", err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

func (c Config) Encrypt(plaintext []byte, recipient string) ([]byte, error) {
	if recipient == "" {
		recipient = c.Recipient
	}
	if recipient == "" {
		return nil, fmt.Errorf("no GPG recipient configured. Run the provider -init option or set GPGRecipient")
	}

	args := c.baseArgs()
	args = append(args, "--trust-model", "always", "--armor", "--recipient", recipient, "--encrypt")

	return c.run(plaintext, args...)
}

func (c Config) Decrypt(ciphertext []byte) ([]byte, error) {
	args := c.baseArgs()
	args = append(args, "--decrypt")

	return c.runWithPassphrase(ciphertext, c.Passphrase, args...)
}

func (c Config) KeyExists(recipient string) (bool, error) {
	if recipient == "" {
		return false, nil
	}
	if err := c.checkExecutable(); err != nil {
		return false, err
	}

	args := c.baseArgs()
	args = append(args, "--with-colons", "--list-keys", recipient)

	cmd := exec.Command(c.execName(), args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check GPG key '%s'. Error: %v", recipient, err)
	}

	return strings.Contains(output.String(), "pub:"), nil
}

func (c Config) Fingerprint(selector string) (string, error) {
	if err := c.checkExecutable(); err != nil {
		return "", err
	}

	args := c.baseArgs()
	args = append(args, "--with-colons", "--fingerprint", selector)

	cmd := exec.Command(c.execName(), args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to read GPG fingerprint for '%s'. Error: %v", selector, err)
	}

	for _, line := range strings.Split(output.String(), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "fpr:") {
			parts := strings.Split(line, ":")
			if len(parts) > 9 {
				fingerprint := strings.TrimSpace(parts[9])
				if fingerprint != "" {
					return fingerprint, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not determine GPG fingerprint for '%s'", selector)
}

func validateKeyFields(name, email, expireDate, keyType, keyLength string) error {
	fields := []struct {
		label string
		value string
	}{
		{label: "name", value: name},
		{label: "email", value: email},
		{label: "expiry", value: expireDate},
		{label: "key type", value: keyType},
		{label: "key length", value: keyLength},
	}
	for _, field := range fields {
		if strings.ContainsAny(field.value, "\r\n\x00") {
			return fmt.Errorf("invalid %s: control characters are not allowed", field.label)
		}
	}
	if !namePattern.MatchString(name) {
		return fmt.Errorf("invalid key name")
	}
	if !emailPattern.MatchString(email) {
		return fmt.Errorf("invalid key email")
	}
	if !expiryPattern.MatchString(expireDate) {
		return fmt.Errorf("invalid key expiry")
	}

	normalizedType := strings.ToUpper(keyType)
	switch normalizedType {
	case "RSA", "DSA", "ECDSA", "ECDH", "EDDSA", "ECC", "ED25519":
	default:
		return fmt.Errorf("invalid key type")
	}
	length, err := strconv.Atoi(keyLength)
	if err != nil {
		return fmt.Errorf("invalid key length")
	}
	if normalizedType == "RSA" || normalizedType == "DSA" {
		if length < 1024 || length > 16384 {
			return fmt.Errorf("invalid key length")
		}
	} else if length < 256 || length > 521 {
		return fmt.Errorf("invalid key length")
	}
	return nil
}

func (c Config) GenerateKey(req KeyRequest) (string, error) {
	if err := c.checkExecutable(); err != nil {
		return "", err
	}

	name := req.Name
	if name == "" {
		name = DefaultName()
	}
	email := req.Email
	if email == "" {
		email = DefaultEmail(name)
	}
	expireDate := req.ExpireDate
	if expireDate == "" {
		expireDate = "0"
	}
	keyType := req.KeyType
	if keyType == "" {
		keyType = "RSA"
	}
	keyLength := req.KeyLength
	if keyLength == "" {
		keyLength = "4096"
	}
	if err := validateKeyFields(name, email, expireDate, keyType, keyLength); err != nil {
		return "", err
	}
	keyType = strings.ToUpper(keyType)

	passphrase := req.Passphrase
	if passphrase == "" {
		passphrase = c.Passphrase
	}

	var batch strings.Builder
	fmt.Fprintf(&batch, "Key-Type: %s\n", keyType)
	fmt.Fprintf(&batch, "Key-Length: %s\n", keyLength)
	fmt.Fprintf(&batch, "Key-Usage: encrypt,sign\n")
	fmt.Fprintf(&batch, "Subkey-Type: %s\n", keyType)
	fmt.Fprintf(&batch, "Subkey-Length: %s\n", keyLength)
	fmt.Fprintf(&batch, "Name-Real: %s\n", name)
	fmt.Fprintf(&batch, "Name-Email: %s\n", email)
	fmt.Fprintf(&batch, "Expire-Date: %s\n", expireDate)
	if passphrase == "" {
		batch.WriteString("%no-protection\n")
	}
	batch.WriteString("%commit\n")

	args := c.baseArgs()
	args = append(args, "--generate-key")

	if _, err := c.runWithPassphrase([]byte(batch.String()), passphrase, args...); err != nil {
		return "", err
	}

	fingerprint, err := c.Fingerprint(email)
	if err != nil {
		return "", err
	}

	return fingerprint, nil
}

func NewMarker(name string, email string, fingerprint string) Marker {
	return Marker{
		Recipient:   fingerprint,
		Fingerprint: fingerprint,
		Name:        name,
		Email:       email,
		Algorithm:   "GPG",
		Created:     time.Now().UTC().Format(time.RFC3339),
	}
}
