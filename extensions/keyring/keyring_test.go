package keyring

import (
	"fmt"
	"testing"
)

func TestSimpleKeyringCreate(t *testing.T) {

	custom := map[string]any{
		"Service": fmt.Sprintf("test_%s", "salainen"),
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create KEYRING storage")
	} else {
		if sto == nil {
			t.Errorf("no KEYRING storage created")
		} else {
			t.Logf("success")
		}
	}

}

func TestSimpleKeyringPutGet(t *testing.T) {

	custom := map[string]any{
		"Prefix": fmt.Sprintf("test_%s", "salainen"),
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create KEYRING storage")
	} else {
		if sto == nil {
			t.Errorf("no KEYRING storage created")
		} else {

			value, err := sto.Get("missingkey")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty: %s", value)
			}

			key := "testcasePutGet"
			secret := "mysecret01"
			err = sto.Put(key, secret)
			if err != nil {
				t.Errorf("failed to put value")
			} else {
				value, err := sto.Get(key)
				if err != nil || value != secret {
					t.Errorf("failed to get value")
				} else {
					t.Logf("success : %s", value)
				}
			}

		}
	}

}
