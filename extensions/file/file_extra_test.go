package file

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/meerkat-manor/salainen"
)

func newTempStorage(t *testing.T, custom map[string]string) salainen.SecretStorage {
	t.Helper()
	settings := map[string]string{"RootPath": t.TempDir()}
	for key, value := range custom {
		settings[key] = value
	}
	storage, err := New("", settings)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}
	return storage
}

func TestFilePipeSuffixIsIgnored(t *testing.T) {

	root := t.TempDir()
	storage, err := New("", map[string]string{"RootPath": root})
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	if err := storage.Put("secret|parameter", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	for _, key := range []string{"secret", "secret|other"} {
		value, err := storage.Get(key)
		if err != nil {
			t.Fatalf("unexpected get error for %q: %v", key, err)
		}
		if value != "value" {
			t.Errorf("expected %q, got %q", "value", value)
		}
	}

	if _, err := os.Stat(filepath.Join(root, "secret")); err != nil {
		t.Errorf("expected file named 'secret' to be created: %v", err)
	}
}

func TestFileNestedPath(t *testing.T) {

	storage := newTempStorage(t, nil)
	key := "one/two/three/secret.txt"

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

func TestFileMissingSecretSentinel(t *testing.T) {

	storage := newTempStorage(t, nil)

	if _, err := storage.Get("does-not-exist"); !errors.Is(err, salainen.ErrNoSuchSecret) {
		t.Errorf("expected ErrNoSuchSecret, got %v", err)
	}
}

func TestFilePutOverwritesExistingValue(t *testing.T) {

	storage := newTempStorage(t, nil)

	if err := storage.Put("key", "first"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}
	if err := storage.Put("key", "second"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	value, err := storage.Get("key")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if value != "second" {
		t.Errorf("expected the latest value, got %q", value)
	}
}

func TestFileProductNameSubstitution(t *testing.T) {

	root := t.TempDir()
	storage, err := New("", map[string]string{
		"RootPath": filepath.Join(root, "{{.ProductName}}"),
	})
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	if err := storage.Put("key", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, salainen.ProductName, "key")); err != nil {
		t.Errorf("expected substituted directory to be used: %v", err)
	}
}

func TestFileRejectsPathTraversal(t *testing.T) {
	storage := newTempStorage(t, nil)

	for _, key := range []string{"../outside", "folder/../../outside", "/absolute"} {
		if err := storage.Put(key, "value"); err == nil {
			t.Errorf("expected put error for %q", key)
		}
		if _, err := storage.Get(key); err == nil {
			t.Errorf("expected get error for %q", key)
		}
	}
}

func TestFileUsesRestrictiveModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not meaningful on Windows")
	}
	root := t.TempDir()
	storage, err := New("", map[string]string{"RootPath": root})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	if err := storage.Put("folder/secret", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	dirInfo, err := os.Stat(filepath.Join(root, "folder"))
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}
	if dirInfo.Mode().Perm()&0o077 != 0 {
		t.Errorf("directory has permissive mode %v", dirInfo.Mode().Perm())
	}

	fileInfo, err := os.Stat(filepath.Join(root, "folder", "secret"))
	if err != nil {
		t.Fatalf("failed to stat secret file: %v", err)
	}
	if fileInfo.Mode().Perm()&0o077 != 0 {
		t.Errorf("secret file has permissive mode %v", fileInfo.Mode().Perm())
	}
}

func TestFileInitWithUnusualCustomValues(t *testing.T) {

	tests := []struct {
		name   string
		custom interface{}
	}{
		{name: "nil", custom: nil},
		{name: "wrong type", custom: "not-a-map"},
		{name: "empty root", custom: map[string]string{"RootPath": ""}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage, err := New("", test.custom)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if storage == nil {
				t.Errorf("expected a storage instance")
			}
		})
	}
}
