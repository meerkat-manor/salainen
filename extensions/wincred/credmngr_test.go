package wincred

import (
	"fmt"
	"testing"
)

func TestSimpleWinCredCreate(t *testing.T) {

	custom := map[string]any{
		"Prefix": fmt.Sprintf("test_%s", "salainen"),
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create ENV storage")
	} else {
		if sto == nil {
			t.Errorf("no ENV storage created")
		} else {
			t.Logf("success")
		}
	}

}

func TestSimpleWinCredPutGet(t *testing.T) {

	custom := map[string]any{
		"Prefix": fmt.Sprintf("test_%s/", "salainen"),
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create ENV storage")
	} else {
		if sto == nil {
			t.Errorf("no ENV storage created")
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
