package promptsec

import (
	"testing"
)

func TestSimplePromptCreate(t *testing.T) {

	sto, err := New("", nil)
	if err != nil {
		t.Errorf("failed to create PROMPT provider")
	} else {
		if sto == nil {
			t.Errorf("no PROMPT provider created")
		} else {
			t.Logf("success")
		}
	}

}

func TestSimplePromptPutGet(t *testing.T) {

	sto, err := New("", nil)
	if err != nil {
		t.Errorf("failed to create PROMPT provider")
	} else {
		if sto == nil {
			t.Errorf("no PROMPT provider created")
		} else {

			value, err := sto.Get("anykey")
			if err != nil {
				// Hard to implement user input in tets framework
				t.Logf("failed to to get prompt")
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
