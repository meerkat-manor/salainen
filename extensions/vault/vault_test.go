package vault

import (
	"testing"

	"github.com/meerkat-manor/salainen/extensions/plain"
)

var internalInit = false

func testInit() {

	if !internalInit {

		configFile := "../../defaults.json"
		err := plain.Register(configFile, nil)
		if err != nil {
			panic(err)
		}
		internalInit = true

	}

}

func TestSimpleVaultCreate(t *testing.T) {

	testInit()
	custom := map[string]string{
		"ApiUrl":      "http://127.0.0.1:8200",
		"AccessToken": "plain:salainen-root",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create VAULT storage. error: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no VAULT storage created")
		} else {
			t.Logf("success")
		}
	}

}

func TestVaultRejectsInsecureRemoteEndpoint(t *testing.T) {
	testInit()
	custom := map[string]string{
		"ApiUrl":      "http://vault.example.com",
		"AccessToken": "plain:salainen-root",
	}
	if _, err := New("", custom); err == nil {
		t.Errorf("expected insecure remote endpoint to be rejected")
	}
}

func TestSimpleVaultPutGet(t *testing.T) {

	addr := requireVault(t)
	testInit()
	custom := map[string]string{
		"ApiUrl":      addr,
		"AccessToken": "plain:salainen-root",
		"Vault":       "secret",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Fatalf("failed to create VAULT storage. error: %v", err)
	}
	if sto == nil {
		t.Fatalf("no VAULT storage created")
	}

	key := "my-secret-password/extra"
	secret := "mysecret01"
	if err := sto.Put(key, secret); err != nil {
		t.Fatalf("failed to put value. error: %v", err)
	}

	value, err := sto.Get(key)
	if err != nil || value != secret {
		t.Errorf("failed to get value: '%s' %v", value, err)
	}

	if value, err := sto.Get("missingkey"); err == nil {
		t.Errorf("failed to flag error for missing key, got value '%s'", value)
	}
}

func TestKV01VaultPutGet(t *testing.T) {

	addr := requireVault(t)
	testInit()
	custom := map[string]string{
		"ApiUrl":      addr,
		"AccessToken": "plain:salainen-root",
		"Vault":       "salainen",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Fatalf("failed to create VAULT storage. error: %v", err)
	}
	if sto == nil {
		t.Fatalf("no VAULT storage created")
	}

	key := "my-kv2-password/extra"
	secret := "mysecret01-99"
	if err := sto.Put(key, secret); err != nil {
		t.Fatalf("failed to put value. error: %v", err)
	}

	value, err := sto.Get(key)
	if err != nil || value != secret {
		t.Errorf("failed to get value: '%s' %v", value, err)
	}

	if value, err := sto.Get("missingkey"); err == nil {
		t.Errorf("failed to flag error for missing key, got value '%s'", value)
	}
}
