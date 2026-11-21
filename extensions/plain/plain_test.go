package plain

import (
	"testing"

	"github.com/meerkat-manor/salainen"
)

func TestSimplePlainCreate(t *testing.T) {

	sto, err := New("", nil)
	if err != nil {
		t.Errorf("failed to create PLAIN storage")
	} else {
		if sto == nil {
			t.Errorf("no PLAIN storage created")
		} else {
			t.Logf("success")
		}
	}

}

func TestSimplePlainPutGet(t *testing.T) {

	sto, err := New("", nil)
	if err != nil {
		t.Errorf("failed to create PLAIN storage")
	} else {
		if sto == nil {
			t.Errorf("no PLAIN storage created")
		} else {

			value, err := sto.Get("anykey")
			if err != nil {
				t.Errorf("failed to to get plain text (itself)")
			} else {
				t.Logf("success with get: %s", value)
			}

			key := "testcasePutGet"
			secret := "mysecret01"
			err = sto.Put(key, secret)
			if err == nil {
				t.Errorf("failed to detect error")
			} else {
				t.Logf("success")
			}

		}
	}

}

func TestPlainRegisterAndHelp(t *testing.T) {

	if err := Register("", nil); err != nil {
		t.Fatalf("failed to register PLAIN storage: %v", err)
	}

	sto, err := salainen.GetSecretStorage("plain")
	if err != nil {
		t.Fatalf("registered PLAIN storage not found: %v", err)
	}

	sto.Help()

	value, err := sto.Get("plain-value")
	if err != nil || value != "plain-value" {
		t.Errorf("unexpected get result: %q %v", value, err)
	}
}
