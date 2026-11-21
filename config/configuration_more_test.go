package config

import (
	"fmt"
	"os"
	"path/filepath"
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
				debugState = false
			})

			dir := t.TempDir()
			path := writeTempFile(t, dir, "config.json", `{"name": "debugvalues"}`)

			if _, err := New(path, false); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if debugState != test.wantDebug {
				t.Errorf("expected debugState %v, got %v", test.wantDebug, debugState)
			}

			_, err := os.Stat(path + ".yaml")
			if test.wantDebug && err != nil {
				t.Errorf("expected YAML copy to be written: %v", err)
			}
			if !test.wantDebug && err == nil {
				t.Errorf("did not expect a YAML copy with debug disabled")
			}
		})
	}
}
