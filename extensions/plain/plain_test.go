package plain

import (
	"testing"
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
