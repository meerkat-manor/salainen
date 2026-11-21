package vault

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/meerkat-manor/salainen"
)

const (
	defaultVaultImage = "hashicorp/vault:1.17"
	defaultVaultToken = "salainen-root"
	testSecretMount   = "salainen"
)

var (
	testVaultAddr      string
	testVaultContainer string
	testVaultErr       error
)

func TestMain(m *testing.M) {
	testVaultAddr, testVaultContainer, testVaultErr = startVaultServer()

	code := m.Run()

	stopVaultServer(testVaultContainer)
	os.Exit(code)
}

func vaultImage() string {
	if image := os.Getenv("SALAINEN_VAULT_IMAGE"); image != "" {
		return image
	}
	return defaultVaultImage
}

func vaultToken() string {
	if token := os.Getenv("SALAINEN_VAULT_TOKEN"); token != "" {
		return token
	}
	return defaultVaultToken
}

func startVaultServer() (string, string, error) {

	if addr := os.Getenv("SALAINEN_VAULT_ADDR"); addr != "" {
		if err := waitForVault(addr, 30*time.Second); err != nil {
			return "", "", err
		}
		return addr, "", ensureSecretMount(addr, vaultToken(), testSecretMount)
	}

	if os.Getenv("SALAINEN_SKIP_DOCKER") != "" {
		return "", "", fmt.Errorf("Docker auto-start disabled by SALAINEN_SKIP_DOCKER")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		return "", "", fmt.Errorf("docker executable not found: %w", err)
	}

	image := vaultImage()
	token := vaultToken()

	args := []string{
		"run", "--detach", "--rm",
		"--cap-add", "IPC_LOCK",
		"-e", "VAULT_DEV_ROOT_TOKEN_ID=" + token,
		"-p", "127.0.0.1::8200",
		image, "server", "-dev",
	}

	output, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("unable to start Vault container from %q: %v: %s",
			image, err, strings.TrimSpace(string(output)))
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return "", "", fmt.Errorf("docker run returned an empty container identifier")
	}

	port, err := waitForPublishedPort(containerID, 3*time.Minute)
	if err != nil {
		_ = stopVaultServer(containerID)
		return "", "", err
	}

	addr := "http://127.0.0.1:" + strconv.Itoa(port)
	if err := waitForVault(addr, 90*time.Second); err != nil {
		_ = stopVaultServer(containerID)
		return "", "", err
	}

	if err := ensureSecretMount(addr, token, testSecretMount); err != nil {
		_ = stopVaultServer(containerID)
		return "", "", err
	}

	return addr, containerID, nil
}

func stopVaultServer(containerID string) error {
	if containerID == "" {
		return nil
	}
	output, err := exec.Command("docker", "rm", "--force", containerID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to remove Vault container %s: %v: %s",
			containerID, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func waitForPublishedPort(containerID string, timeout time.Duration) (int, error) {

	deadline := time.Now().Add(timeout)
	lastErr := error(nil)

	for time.Now().Before(deadline) {
		output, err := exec.Command("docker", "port", containerID, "8200/tcp").CombinedOutput()
		if err != nil {
			lastErr = err
		} else {
			mapping := strings.TrimSpace(string(output))
			if idx := strings.LastIndex(mapping, ":"); idx >= 0 && idx < len(mapping)-1 {
				if port, err := strconv.Atoi(mapping[idx+1:]); err == nil {
					return port, nil
				}
			}
			lastErr = fmt.Errorf("unexpected port mapping %q", mapping)
		}
		time.Sleep(500 * time.Millisecond)
	}

	return 0, fmt.Errorf("timed out waiting for Vault port mapping: %v", lastErr)
}

func waitForVault(addr string, timeout time.Duration) error {

	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)
	lastErr := error(nil)

	for time.Now().Before(deadline) {
		response, err := client.Get(addr + "/v1/sys/health")
		if err != nil {
			lastErr = err
		} else {
			status := response.StatusCode
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			switch status {
			case http.StatusOK, http.StatusTooManyRequests, 472, 473, 501:
				return nil
			default:
				lastErr = fmt.Errorf("health endpoint returned status %d", status)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("Vault not ready at %s: %v", addr, lastErr)
}

func ensureSecretMount(addr, token, path string) error {

	request, err := http.NewRequest(http.MethodPost, addr+"/v1/sys/mounts/"+path,
		strings.NewReader(`{"type":"kv","options":{"version":"2"}}`))
	if err != nil {
		return err
	}
	request.Header.Set("X-Vault-Token", token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("unable to enable KV v2 mount %q: %w", path, err)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}
	if response.StatusCode == http.StatusBadRequest && strings.Contains(string(body), "path is already in use") {
		return nil
	}

	return fmt.Errorf("unable to enable KV v2 mount %q: status %d: %s", path, response.StatusCode, strings.TrimSpace(string(body)))
}

func requireVault(t *testing.T) string {
	t.Helper()
	if testVaultErr != nil {
		t.Skipf("Vault integration server unavailable: %v", testVaultErr)
	}
	return testVaultAddr
}

func newTestStorage(t *testing.T, addr string, extra map[string]string) salainen.SecretStorage {
	t.Helper()
	testInit()

	custom := map[string]string{
		"ApiUrl":      addr,
		"AccessToken": "plain:" + vaultToken(),
		"Vault":       testSecretMount,
	}
	for key, value := range extra {
		custom[key] = value
	}

	storage, err := New("", custom)
	if err != nil {
		t.Fatalf("failed to create Vault storage: %v", err)
	}
	return storage
}

func TestVaultPutOverwrite(t *testing.T) {

	addr := requireVault(t)
	storage := newTestStorage(t, addr, nil)

	if err := storage.Put("overwrite/key", "first"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}
	if err := storage.Put("overwrite/key", "second"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	value, err := storage.Get("overwrite/key")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if value != "second" {
		t.Errorf("expected the latest value, got %q", value)
	}
}

func TestVaultPipeSuffixIsIgnored(t *testing.T) {

	addr := requireVault(t)
	storage := newTestStorage(t, addr, nil)

	if err := storage.Put("with-suffix|ignored", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	value, err := storage.Get("with-suffix")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if value != "value" {
		t.Errorf("expected %q, got %q", "value", value)
	}
}

func TestVaultNestedPath(t *testing.T) {

	addr := requireVault(t)
	storage := newTestStorage(t, addr, nil)

	const key = "nested/level/two/secret"
	if err := storage.Put(key, "nested-value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	value, err := storage.Get(key)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if value != "nested-value" {
		t.Errorf("unexpected value: %q", value)
	}
}

func TestVaultMissingSecret(t *testing.T) {

	addr := requireVault(t)
	storage := newTestStorage(t, addr, nil)

	if err := storage.Put("existing-key", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	if value, err := storage.Get("definitely-missing"); err == nil {
		t.Errorf("expected an error for a missing secret, got value %q", value)
	}
}

func TestVaultCustomElementName(t *testing.T) {

	addr := requireVault(t)
	storage := newTestStorage(t, addr, map[string]string{"ElementName": "passphrase"})

	if err := storage.Put("custom-element", "custom-value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	value, err := storage.Get("custom-element")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if value != "custom-value" {
		t.Errorf("unexpected value: %q", value)
	}

	defaultStorage := newTestStorage(t, addr, nil)
	if err := defaultStorage.Put("default-element-key", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}
	if _, err := defaultStorage.Get("custom-element"); err == nil {
		t.Errorf("expected an error when reading a secret with a different element name")
	}
}

func TestVaultDefaultMount(t *testing.T) {

	addr := requireVault(t)
	storage := newTestStorage(t, addr, map[string]string{"Vault": ""})

	if err := storage.Put("default-mount/key", "default-value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	value, err := storage.Get("default-mount/key")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if value != "default-value" {
		t.Errorf("unexpected value: %q", value)
	}
}

func TestVaultInvalidToken(t *testing.T) {

	addr := requireVault(t)
	testInit()

	storage, err := New("", map[string]string{
		"ApiUrl":      addr,
		"AccessToken": "plain:not-a-valid-token",
		"Vault":       testSecretMount,
	})
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	if err := storage.Put("invalid-token/key", "value"); err == nil {
		t.Errorf("expected an authentication error on put")
	}
	if _, err := storage.Get("invalid-token/key"); err == nil {
		t.Errorf("expected an authentication error on get")
	}
}

func TestVaultInitValidation(t *testing.T) {

	testInit()

	tests := []struct {
		name   string
		custom map[string]string
	}{
		{name: "no ApiUrl", custom: map[string]string{"AccessToken": "plain:token"}},
		{name: "no AccessToken", custom: map[string]string{"ApiUrl": "http://127.0.0.1:8200"}},
		{name: "unresolvable token", custom: map[string]string{
			"ApiUrl": "http://127.0.0.1:8200", "AccessToken": "missingprovider:token"}},
		{name: "empty plain token reference", custom: map[string]string{
			"ApiUrl": "http://127.0.0.1:8200", "AccessToken": "plain:"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := New("", test.custom); err == nil {
				t.Errorf("expected an error")
			}
		})
	}

	if _, err := New("", map[string]string{
		"ApiUrl":      "http://127.0.0.1:8200",
		"AccessToken": "vault:token",
	}); err == nil {
		t.Errorf("expected an error for a looping provider reference")
	}
}

func TestVaultRegister(t *testing.T) {

	testInit()

	if err := Register("", map[string]string{
		"ApiUrl":      "http://127.0.0.1:8200",
		"AccessToken": "plain:" + vaultToken(),
	}); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	t.Cleanup(func() {
		salainen.RemoveSecretStorage("vault")
	})

	storage, err := salainen.GetSecretStorage("vault")
	if err != nil {
		t.Errorf("expected the vault provider to be registered: %v", err)
	}
	if storage == nil {
		t.Errorf("expected a storage instance")
	}
}
