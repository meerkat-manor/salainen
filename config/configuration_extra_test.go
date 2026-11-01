package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meerkat-manor/salainen"
)

func writeTempFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	return path
}

func TestLoadConfigJSON(t *testing.T) {

	dir := t.TempDir()
	path := writeTempFile(t, dir, "config.json", `{
		"name": "testapp",
		"version": "1.2.3",
		"providers": {
			"plain": {
				"enabled": true,
				"name": "Plain text",
				"custom": {
					"Key": "Value"
				}
			}
		}
	}`)

	conf, err := loadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conf.Name != "testapp" {
		t.Errorf("unexpected name: %q", conf.Name)
	}
	if conf.Version != "1.2.3" {
		t.Errorf("unexpected version: %q", conf.Version)
	}

	storage, exists := conf.Storage["plain"]
	if !exists {
		t.Fatalf("plain provider not loaded")
	}
	if !storage.Enabled {
		t.Errorf("plain provider not enabled")
	}
	if storage.Name != "Plain text" {
		t.Errorf("unexpected provider name: %q", storage.Name)
	}
	if storage.Custom["Key"] != "Value" {
		t.Errorf("unexpected custom value: %q", storage.Custom["Key"])
	}
}

func TestLoadConfigYAML(t *testing.T) {

	dir := t.TempDir()
	path := writeTempFile(t, dir, "config.yaml", `
name: yamlapp
version: "0.1.0"
providers:
  env:
    enabled: true
    name: Environment
    custom:
      Prefix: test_
`)

	conf, err := loadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conf.Name != "yamlapp" {
		t.Errorf("unexpected name: %q", conf.Name)
	}
	if conf.Version != "0.1.0" {
		t.Errorf("unexpected version: %q", conf.Version)
	}

	storage, exists := conf.Storage["env"]
	if !exists {
		t.Fatalf("env provider not loaded")
	}
	if storage.Custom["Prefix"] != "test_" {
		t.Errorf("unexpected custom prefix: %q", storage.Custom["Prefix"])
	}
}

func TestLoadConfigYmlExtension(t *testing.T) {

	dir := t.TempDir()
	path := writeTempFile(t, dir, "config.yml", "name: ymlapp\nversion: \"2.0.0\"\n")

	conf, err := loadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.Name != "ymlapp" || conf.Version != "2.0.0" {
		t.Errorf("unexpected configuration: %q %q", conf.Name, conf.Version)
	}
}

func TestLoadConfigErrors(t *testing.T) {

	dir := t.TempDir()

	t.Run("missing file", func(t *testing.T) {
		if _, err := loadConfig(filepath.Join(dir, "missing.json")); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("directory", func(t *testing.T) {
		if _, err := loadConfig(dir); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		path := writeTempFile(t, dir, "invalid.json", "{ not json")
		if _, err := loadConfig(path); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		path := writeTempFile(t, dir, "invalid.yaml", "name: [unclosed")
		if _, err := loadConfig(path); err == nil {
			t.Errorf("expected error")
		}
	})
}

func TestCheckConfigurationFile(t *testing.T) {

	empty := t.TempDir()
	if got := checkConfigurationFile(empty); got != "" {
		t.Errorf("expected no configuration file, got %q", got)
	}

	for _, name := range []string{"salainen.json", "salainen.yaml", "salainen.yml"} {
		dir := t.TempDir()
		path := writeTempFile(t, dir, name, "")
		if got := checkConfigurationFile(dir); got != path {
			t.Errorf("expected %q, got %q", path, got)
		}
	}
}

func TestNewWithMissingOrInvalidConfigFile(t *testing.T) {

	dir := t.TempDir()

	if _, err := New(filepath.Join(dir, "missing.json"), false); err == nil {
		t.Errorf("expected error for missing configuration file")
	}

	if _, err := New(dir, false); err == nil {
		t.Errorf("expected error for configuration directory")
	}
}

func TestNewWithHomeRelativePath(t *testing.T) {

	dir := t.TempDir()
	writeTempFile(t, dir, "salainen.json", `{"name": "homeapp", "version": "1.0.0"}`)

	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOME", dir)
	t.Setenv(strings.ToUpper(os.Getenv(strings.ToUpper(salainen.ProductName)+"_DEBUG")), "")

	app, err := New("~/salainen.json", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Name != "homeapp" {
		t.Errorf("unexpected app name: %q", app.Name)
	}
}

func TestNewFromEnvironment(t *testing.T) {

	dir := t.TempDir()
	path := writeTempFile(t, dir, "envconfig.json", `{
		"name": "envapp",
		"version": "3.0.0",
		"providers": {
			"plain": {"enabled": true, "name": "Plain"}
		}
	}`)

	t.Setenv(ENV_SALAINEN, path)
	t.Setenv(ENV_SALAINEN_DEBUG, "")

	app, err := New("", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Name != "envapp" {
		t.Errorf("unexpected app name: %q", app.Name)
	}
	if app.StorageName["plain"] != "Plain" {
		t.Errorf("plain provider not registered")
	}
}

func TestNewWithUnknownProvider(t *testing.T) {

	dir := t.TempDir()
	path := writeTempFile(t, dir, "unknown.json", `{
		"providers": {
			"notaprovider": {"enabled": true}
		}
	}`)

	if _, err := New(path, false); err == nil {
		t.Errorf("expected error for unknown provider")
	}

	if _, err := New(path, true); err == nil {
		t.Errorf("expected error for unknown provider even when ignoring provider errors")
	}
}

func TestNewIgnoresProviderErrors(t *testing.T) {

	dir := t.TempDir()
	path := writeTempFile(t, dir, "badprovider.json", `{
		"name": "badprovider",
		"providers": {
			"efile": {
				"enabled": true,
				"name": "Encrypted File",
				"custom": {"Algorithm": "not-an-algorithm"}
			}
		}
	}`)

	if _, err := New(path, false); err == nil {
		t.Errorf("expected provider error")
	}

	app, err := New(path, true)
	if err != nil {
		t.Fatalf("unexpected error when ignoring provider errors: %v", err)
	}
	if app.StorageName["efile"] != "Encrypted File" {
		t.Errorf("efile provider not recorded")
	}
}

func TestNewWithDefaults(t *testing.T) {

	dir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	t.Setenv(ENV_SALAINEN, "")
	t.Setenv(ENV_SALAINEN_DEBUG, "")
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOME", dir)

	app, err := New("", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, name := range []string{"plain", "env", "file", "efile", "keyring", "prompt"} {
		if _, exists := app.StorageName[name]; !exists {
			t.Errorf("default provider %q not registered", name)
		}
	}
}

func TestDebugConfigurationWritesYAML(t *testing.T) {

	dir := t.TempDir()
	path := writeTempFile(t, dir, "debug.json", `{"name": "debugapp", "version": "1.0.0"}`)

	t.Setenv(strings.ToUpper(os.Getenv(strings.ToUpper(salainen.ProductName)+"_DEBUG")), "TRUE")
	t.Cleanup(func() {
		debugState = false
	})

	if _, err := New(path, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path + ".yaml"); err != nil {
		t.Errorf("expected YAML copy to be written: %v", err)
	}
}
