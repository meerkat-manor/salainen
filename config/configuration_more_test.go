package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadConfigExtensionHandling(t *testing.T) {

	dir := t.TempDir()

	t.Run("uppercase json extension", func(t *testing.T) {
		path := writeTempFile(t, dir, "uppercase.JSON", `{"name": "uppercase"}`)
		conf, err := loadConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if conf.Name != "uppercase" {
			t.Errorf("unexpected name: %q", conf.Name)
		}
	})

	t.Run("yaml content with unknown extension", func(t *testing.T) {
		path := writeTempFile(t, dir, "unknown.txt", "name: fallback\nversion: \"9.9\"\n")
		conf, err := loadConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if conf.Name != "fallback" || conf.Version != "9.9" {
			t.Errorf("unexpected configuration: %q %q", conf.Name, conf.Version)
		}
	})

	t.Run("json content with yaml extension", func(t *testing.T) {
		path := writeTempFile(t, dir, "mismatch.yaml", `{"name": "mismatch"}`)
		conf, err := loadConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if conf.Name != "mismatch" {
			t.Errorf("unexpected name: %q", conf.Name)
		}
	})
}

func TestCheckConfigurationFilePriority(t *testing.T) {

	dir := t.TempDir()
	for _, name := range []string{"salainen.yml", "salainen.yaml", fmt.Sprintf("%s.json", ProductName)} {
		writeTempFile(t, dir, name, "")
	}

	want := filepath.Join(dir, fmt.Sprintf("%s.json", ProductName))
	if got := checkConfigurationFile(dir); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCheckConfigurationFileSkipsDirectories(t *testing.T) {

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, fmt.Sprintf("%s.json", ProductName)), 0o700); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	want := writeTempFile(t, dir, "salainen.yaml", "name: test\n")

	if got := checkConfigurationFile(dir); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCheckConfigurationFileSkipsWorldWritable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not meaningful on Windows")
	}

	dir := t.TempDir()
	path := writeTempFile(t, dir, "salainen.json", `{"name":"insecure"}`)
	if err := os.Chmod(path, 0o666); err != nil {
		t.Fatalf("failed to loosen file permissions: %v", err)
	}
	if got := checkConfigurationFile(dir); got == path {
		t.Errorf("world-writable configuration file was accepted")
	}
}

func TestDebugFlagValues(t *testing.T) {

	tests := []struct {
		name      string
		value     string
		wantDebug bool
	}{
		{name: "TRUE", value: "TRUE", wantDebug: true},
		{name: "lowercase true", value: "true", wantDebug: true},
		{name: "YES", value: "YES", wantDebug: true},
		{name: "one", value: "1", wantDebug: true},
		{name: "false", value: "false", wantDebug: false},
		{name: "empty", value: "", wantDebug: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(ENV_SALAINEN_DEBUG, test.value)
			t.Cleanup(func() {
				debugState.Store(false)
			})

			dir := t.TempDir()
			path := writeTempFile(t, dir, "config.json", `{"name": "debugvalues"}`)

			if _, err := New(path, false); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := debugState.Load(); got != test.wantDebug {
				t.Errorf("expected debugState %v, got %v", test.wantDebug, got)
			}

			if _, err := os.Stat(path + ".yaml"); err == nil {
				t.Errorf("did not expect a YAML copy to be written")
			}

		})
	}
}
