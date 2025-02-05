package encryptedfile

import (
	"fmt"
	"testing"
)

func TestSimpleEncFileCreate(t *testing.T) {

	custom := map[string]any{
		"RootPath":  fmt.Sprintf("~/.secrets/test_%s", "salainen"),
		"Algorithm": "",
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

func TestSimpleEncFilePutGet(t *testing.T) {

	custom := map[string]any{
		"RootPath":  fmt.Sprintf("~/.secrets/test_%s", "salainen"),
		"Algorithm": "",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create file storage: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no file storage created")
		} else {

			value, err := sto.Get("missingkey")
			if err == nil {
				t.Errorf("failed to flag error: %v", err)
			} else {
				t.Logf("success with empty: %s", value)
			}

			key := "testcaseEncPutGet.txt"
			secret := "mysecretenc01"
			err = sto.Put(key, secret)
			if err != nil {
				t.Errorf("failed to put value: %v", err)
			} else {
				value, err := sto.Get(key)
				if err != nil || value != secret {
					t.Errorf("failed to get value: '%s' %v", value, err)
				} else {
					t.Logf("success : %s", value)
				}
			}

		}
	}

}

func TestPassEncFilePutGet(t *testing.T) {

	custom := map[string]any{
		"RootPath":  fmt.Sprintf("~/.secrets/test_%s", "salainen"),
		"Algorithm": "",
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
				t.Errorf("failed to flag error: %v", err)
			} else {
				t.Logf("success with empty: %s", value)
			}

			key := "testcaseEncPutGet.txt|hty5784"
			secret := "mysecretenc01"
			err = sto.Put(key, secret)
			if err != nil {
				t.Errorf("failed to put value")
			} else {
				value, err := sto.Get(key)
				if err != nil || value != secret {
					t.Errorf("part 1 failed to get value")
				} else {
					t.Logf("part 1 success : %s", value)
				}

				// Now try GET with different "pass"
				key = "testcaseEncPutGet.txt|hfyuiety5784"
				value, err = sto.Get(key)
				if err != nil || value != secret {
					t.Logf("part 2 expected fail : %s", value)
				} else {
					t.Errorf("part 2 unexpected success: %s", value)
				}

				// Now try GET with NO "pass"
				key = "testcaseEncPutGet.txt|"
				value, err = sto.Get(key)
				if err != nil || value != secret {
					t.Logf("part 3 expected fail : %s", value)
				} else {
					t.Errorf("part 3 unexpected success: %s", value)
				}
			}

		}
	}

}

func TestUnknownEncFileCreate(t *testing.T) {

	custom := map[string]any{
		"RootPath":  fmt.Sprintf("~/.secrets/test_%s", "salainen"),
		"Algorithm": "xyz",
	}

	_, err := New("", custom)
	if err != nil {
		t.Logf("success")
	} else {
		t.Errorf("failed to detect invalid algorithm")
	}

}
