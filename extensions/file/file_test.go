package file

import (
	"fmt"
	"testing"
)

func TestSimpleFileCreate(t *testing.T) {

	custom := map[string]any{
		"RootPath": fmt.Sprintf("~/.secrets/test_%s", "salainen"),
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create file storage")
	} else {
		if sto == nil {
			t.Errorf("no file storage created")
		} else {
			t.Logf("success")
		}
	}

}

func TestSimpleFilePutGet(t *testing.T) {

	custom := map[string]any{
		"RootPath": fmt.Sprintf("~/.secrets/test_%s", "salainen"),
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create file storage")
	} else {
		if sto == nil {
			t.Errorf("no file storage created")
		} else {

			value, err := sto.Get("missingkey")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty: %s", value)
			}

			key := "testcasePutGet.txt"
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
