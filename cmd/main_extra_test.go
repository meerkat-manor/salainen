package main

import (
	"path/filepath"
	"testing"

	"github.com/meerkat-manor/salainen"
)

type cmdFakeStorage struct {
	values map[string]string
	getErr error
	putErr error
}

func newCmdFakeStorage() *cmdFakeStorage {
	return &cmdFakeStorage{values: map[string]string{}}
}

func (f *cmdFakeStorage) Help() {}

func (f *cmdFakeStorage) Init(custom interface{}) error {
	return nil
}

func (f *cmdFakeStorage) Get(path string) (string, error) {
	if f.getErr != nil {
		return "", f.getErr
	}
	value, exists := f.values[path]
	if !exists {
		return "", salainen.ErrNoSuchSecret
	}
	return value, nil
}

func (f *cmdFakeStorage) Put(path string, value string) error {
	if f.putErr != nil {
		return f.putErr
	}
	f.values[path] = value
	return nil
}

func registerCmdFakeStorage(t *testing.T) *cmdFakeStorage {
	t.Helper()
	storage := newCmdFakeStorage()
	salainen.AddSecretStorage("cmdtest", storage)
	t.Cleanup(func() {
		delete(salainen.SecretStorages, "cmdtest")
	})
	return storage
}

func TestProcessDefaultParameterErrors(t *testing.T) {

	if err := process_default(false, false, true, []string{}); err == nil {
		t.Errorf("expected error with no arguments")
	}

	if err := process_default(false, false, true, []string{"a", "b", "c"}); err == nil {
		t.Errorf("expected error with too many arguments")
	}

	if err := process_default(false, true, true, []string{"cmdtest:key", "value"}); err == nil {
		t.Errorf("expected error when generating with a supplied value")
	}
}

func TestProcessDefaultGetPut(t *testing.T) {

	storage := registerCmdFakeStorage(t)

	if err := process_default(false, false, false, []string{"cmdtest:key", "first"}); err != nil {
		t.Fatalf("unexpected error saving secret: %v", err)
	}
	if storage.values["key"] != "first" {
		t.Errorf("secret not saved")
	}

	if err := process_default(false, false, false, []string{"cmdtest:key", "second"}); err == nil {
		t.Errorf("expected error saving over existing secret without overwrite")
	}
	if storage.values["key"] != "first" {
		t.Errorf("secret should not have been overwritten")
	}

	if err := process_default(false, false, true, []string{"cmdtest:key", "second"}); err != nil {
		t.Fatalf("unexpected error overwriting secret: %v", err)
	}
	if storage.values["key"] != "second" {
		t.Errorf("secret not overwritten")
	}

	if err := process_default(false, false, true, []string{"cmdtest:key"}); err != nil {
		t.Errorf("unexpected error reading secret: %v", err)
	}

	if err := process_default(false, false, true, []string{"cmdtest:missing"}); err == nil {
		t.Errorf("expected error reading missing secret")
	}
}

func TestProcessDefaultGetError(t *testing.T) {

	storage := registerCmdFakeStorage(t)
	storage.getErr = salainen.ErrInvalidSecretAccess

	if err := process_default(false, false, true, []string{"cmdtest:key"}); err == nil {
		t.Errorf("expected error from storage")
	}
}

func TestProcessDefaultPutError(t *testing.T) {

	storage := registerCmdFakeStorage(t)
	storage.putErr = salainen.ErrUnableToSave

	if err := process_default(false, false, true, []string{"cmdtest:key", "value"}); err == nil {
		t.Errorf("expected error from storage")
	}
}

func TestProcessDefaultGenerate(t *testing.T) {

	storage := registerCmdFakeStorage(t)

	if err := process_default(false, true, false, []string{"cmdtest:generated"}); err != nil {
		t.Fatalf("unexpected error generating secret: %v", err)
	}

	value, exists := storage.values["generated"]
	if !exists {
		t.Fatalf("generated secret not saved")
	}
	if len(value) < 10 || len(value) > 20 {
		t.Errorf("unexpected generated secret length: %d", len(value))
	}
}

func TestProcessStorageCaseInsensitive(t *testing.T) {

	configFile := "../defaults.json"
	if err := process_storage(&configFile, []string{"ENV"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := process_storage(&configFile, []string{"plain"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcessStorageMissingConfig(t *testing.T) {

	configFile := filepath.Join(t.TempDir(), "missing.json")
	if err := process_storage(&configFile, []string{"env"}); err == nil {
		t.Errorf("expected error for missing configuration file")
	}
}

func TestPrintUsage(t *testing.T) {

	PrintUsage()
}

func TestProviderStorageTypes(t *testing.T) {

	configFile := "../defaults.json"
	ProviderStorageTypes(&configFile)
}
