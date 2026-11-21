package salainen

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestGetWithEmptyKeyPath(t *testing.T) {

	storage := &fakeStorage{getValue: "should-not-be-read"}
	registerFake(t, "emptykeypath", storage)

	value, err := Get("emptykeypath:")
	if err == nil {
		t.Fatalf("expected error for empty key path")
	}
	if !strings.Contains(err.Error(), "missing secret key path") {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "" {
		t.Errorf("expected empty value, got %q", value)
	}
	if storage.lastGetPath != "" {
		t.Errorf("storage should not have been called, got path %q", storage.lastGetPath)
	}
}

func TestSearchSecretStorageDefaultVariants(t *testing.T) {

	registerFake(t, "env", &fakeStorage{})

	tests := map[string]string{
		"$env:a:b":   "a:b",
		"${env:a:b}": "a:b",
		"${a:b}":     "a:b",
		"$env:":      "",
		"${env:}":    "",
		"${}":        "",
	}

	for id, expected := range tests {
		t.Run(id, func(t *testing.T) {
			_, path, err := SearchSecretStorage(id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if path != expected {
				t.Errorf("expected path %q, got %q", expected, path)
			}
		})
	}
}

func TestSearchSecretStorageMaximumLengthBoundary(t *testing.T) {

	storage := &fakeStorage{}
	registerFake(t, "boundary", storage)

	maxLen := MaxKeyPathLength + MaxProviderLength + 1
	padding := maxLen - len("boundary:")
	id := "boundary:" + strings.Repeat("k", padding)

	if _, path, err := SearchSecretStorage(id); err != nil {
		t.Errorf("unexpected error at maximum length: %v", err)
	} else if path != strings.Repeat("k", padding) {
		t.Errorf("unexpected path at maximum length")
	}

	if _, _, err := SearchSecretStorage(id + "x"); err == nil {
		t.Errorf("expected error above maximum length")
	}
}

func TestAddSecretStorageOverwritesRegistration(t *testing.T) {

	first := &fakeStorage{getValue: "first"}
	registerFake(t, "replaceable", first)

	second := &fakeStorage{getValue: "second"}
	registerFake(t, "replaceable", second)

	got, err := GetSecretStorage("replaceable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != second {
		t.Errorf("expected the most recent registration to be returned")
	}

	value, err := Get("replaceable:any")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "second" {
		t.Errorf("expected value from replacement storage, got %q", value)
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	const workers = 8
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		name := fmt.Sprintf("concurrentracer%d", i)
		t.Cleanup(func() {
			RemoveSecretStorage(name)
		})

		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				AddSecretStorage(name, &fakeStorage{getValue: name})
				ListSecretStorage()
				GetSecretStorage(name)
				SearchSecretStorage(name + ":key")
				RemoveSecretStorage(name)
			}
		}(name)
	}

	wg.Wait()
}

func TestPutAtMaximumValueLength(t *testing.T) {

	storage := &fakeStorage{}
	registerFake(t, "maxvalue", storage)

	value := strings.Repeat("v", MaxValueLength)
	if err := Put("maxvalue:key", value); err != nil {
		t.Errorf("unexpected error at maximum value length: %v", err)
	}
	if storage.lastPutVal != value {
		t.Errorf("value was not stored in full")
	}
}
