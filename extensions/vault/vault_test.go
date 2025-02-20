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

func TestSimpleVaultPutGet(t *testing.T) {

	testInit()
	custom := map[string]string{
		"ApiUrl":      "http://127.0.0.1:8200",
		"AccessToken": "plain:salainen-root",
		"Vault":       "secret",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create VAULT storage. error: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no VAULT storage created")
		} else {

			value, err := sto.Get("missingkey")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty: '%s' and %v", value, err)
			}

			key := "my-secret-password/extra"
			secret := "mysecret01"
			err = sto.Put(key, secret)
			if err != nil {
				t.Errorf("failed to put value. error: %v", err)
			} else {
				value, err := sto.Get(key)
				if err != nil || value != secret {
					t.Errorf("failed to get value. error: %v", err)
				} else {
					t.Logf("success : %s", value)
				}
			}

		}
	}

}

func TestKV01VaultPutGet(t *testing.T) {

	testInit()
	custom := map[string]string{
		"ApiUrl":      "http://127.0.0.1:8200",
		"AccessToken": "plain:salainen-root",
		"Vault":       "salainen",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create VAULT storage. error: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no VAULT storage created")
		} else {

			value, err := sto.Get("missingkey")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty: '%s' and %v", value, err)
			}

			key := "my-kv2-password/extra"
			secret := "mysecret01-99"
			err = sto.Put(key, secret)
			if err != nil {
				t.Errorf("failed to put value. error: %v", err)
			} else {
				value, err := sto.Get(key)
				if err != nil || value != secret {
					t.Errorf("failed to get value. error: %v", err)
				} else {
					t.Logf("success : %s", value)
				}
			}

		}
	}

}
