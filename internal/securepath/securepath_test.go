package securepath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveWithNonexistentRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "secrets")
	got, err := Resolve(root, "folder/key.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, "folder", "key.txt")
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestResolveRejectsUnsafePaths(t *testing.T) {
	tests := []string{
		"../outside",
		"folder/../../outside",
		`folder\..\outside`,
		"/absolute/path",
		`C:\absolute\path`,
	}
	for _, key := range tests {
		t.Run(key, func(t *testing.T) {
			if _, err := Resolve(t.TempDir(), key); err == nil {
				t.Errorf("expected error for %q", key)
			}
		})
	}
}

func TestResolveExpandsTildeWithoutRoot(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("home directory unavailable: %v", err)
	}
	got, err := Resolve("", "~/salainen-test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, "salainen-test-key")
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
