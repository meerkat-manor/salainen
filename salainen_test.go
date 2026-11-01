package salainen

import (
	"fmt"
	"strings"
	"testing"
)

type fakeStorage struct {
	getValue string
	getErr   error
	putErr   error

	lastGetPath string
	lastPutPath string
	lastPutVal  string
	helpCalled  bool
}

func (f *fakeStorage) Help() {
	f.helpCalled = true
}

func (f *fakeStorage) Init(custom interface{}) error {
	return nil
}

func (f *fakeStorage) Get(path string) (string, error) {
	f.lastGetPath = path
	return f.getValue, f.getErr
}

func (f *fakeStorage) Put(path string, value string) error {
	f.lastPutPath = path
	f.lastPutVal = value
	return f.putErr
}

func registerFake(t *testing.T, name string, storage *fakeStorage) {
	t.Helper()
	AddSecretStorage(name, storage)
	t.Cleanup(func() {
		delete(SecretStorages, name)
	})
}

func TestGet(t *testing.T) {

	if _, err := Get(""); err == nil {
		t.Errorf("expected error for empty key")
	}

	if _, err := Get("noseparator"); err == nil {
		t.Errorf("expected error for key without provider separator")
	}

	if _, err := Get("missingprovider:some/path"); err == nil {
		t.Errorf("expected error for unknown provider")
	}

	storage := &fakeStorage{getValue: "secret-value"}
	registerFake(t, "fakeget", storage)

	value, err := Get("fakeget:some/path")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "secret-value" {
		t.Errorf("unexpected value: %q", value)
	}
	if storage.lastGetPath != "some/path" {
		t.Errorf("unexpected path: %q", storage.lastGetPath)
	}

	storage.getErr = fmt.Errorf("lookup failed")
	if _, err := Get("fakeget:some/path"); err == nil {
		t.Errorf("expected storage error to be returned")
	}
}

func TestPut(t *testing.T) {

	if err := Put("", "value"); err == nil {
		t.Errorf("expected error for empty key")
	}

	if err := Put("fakeget:some/path", ""); err == nil {
		t.Errorf("expected error for empty value")
	}

	tooLong := strings.Repeat("a", MaxValueLength+1)
	if err := Put("fakeget:some/path", tooLong); err == nil {
		t.Errorf("expected error for oversized value")
	}

	if err := Put("missingprovider:some/path", "value"); err == nil {
		t.Errorf("expected error for unknown provider")
	}

	storage := &fakeStorage{}
	registerFake(t, "fakeput", storage)

	if err := Put("fakeput:some/path", "value"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if storage.lastPutPath != "some/path" || storage.lastPutVal != "value" {
		t.Errorf("unexpected put call: %q %q", storage.lastPutPath, storage.lastPutVal)
	}

	storage.putErr = fmt.Errorf("write failed")
	if err := Put("fakeput:some/path", "value"); err == nil {
		t.Errorf("expected storage error to be returned")
	}
}

func TestAddAndListSecretStorage(t *testing.T) {

	storage := &fakeStorage{}
	registerFake(t, "fakelist", storage)

	list := ListSecretStorage()
	if _, exists := list["fakelist"]; !exists {
		t.Errorf("registered storage not listed")
	}

	got, err := GetSecretStorage("fakelist")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got != storage {
		t.Errorf("unexpected storage returned")
	}

	if _, err := GetSecretStorage("doesnotexist"); err != ErrInvalidProvider {
		t.Errorf("expected ErrInvalidProvider, got %v", err)
	}
}

func TestSearchSecretStorage(t *testing.T) {

	if _, _, err := SearchSecretStorage(""); err == nil {
		t.Errorf("expected error for empty identifier")
	}

	maxLen := MaxKeyPathLength + MaxProviderLength + 1
	if _, _, err := SearchSecretStorage(strings.Repeat("a", maxLen+1)); err == nil {
		t.Errorf("expected error for oversized identifier")
	}

	if _, _, err := SearchSecretStorage("noseparator"); err != ErrInvalidProvider {
		t.Errorf("expected ErrInvalidProvider, got %v", err)
	}

	if _, _, err := SearchSecretStorage(":path"); err != ErrInvalidProvider {
		t.Errorf("expected ErrInvalidProvider for empty provider, got %v", err)
	}

	if _, _, err := SearchSecretStorage("missingprovider:path"); err != ErrInvalidProvider {
		t.Errorf("expected ErrInvalidProvider for unknown provider, got %v", err)
	}

	storage := &fakeStorage{}
	registerFake(t, "fakesearch", storage)
	registerFake(t, "env", &fakeStorage{})

	got, path, err := SearchSecretStorage("fakesearch:key/path")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got != storage {
		t.Errorf("unexpected storage returned")
	}
	if path != "key/path" {
		t.Errorf("unexpected path: %q", path)
	}

	defaults := map[string]string{
		"$env:MY_VAR":     "MY_VAR",
		"${env:MY_VAR}":   "MY_VAR",
		"${MY_VAR}":       "MY_VAR",
		"fakesearch:a:b":  "a:b",
		"fakesearch:path": "path",
	}

	for id, expected := range defaults {
		_, path, err := SearchSecretStorage(id)
		if err != nil {
			t.Errorf("unexpected error for %q: %v", id, err)
			continue
		}
		if path != expected {
			t.Errorf("for %q expected path %q, got %q", id, expected, path)
		}
	}
}
