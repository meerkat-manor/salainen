package keepass

import (
	"fmt"
	"os"
	"testing"
)

func TestSimpleKPCreate(t *testing.T) {

	custom := map[string]any{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "tester01",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create KEEPASS storage: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no KEEPASS storage created")
		} else {
			t.Logf("success")
		}
	}

}

func TestSimpleKPPutGet(t *testing.T) {

	custom := map[string]any{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "tester01",
	}
	//custom["Path"] = `C:\tmp\salainen\tests\data\test_secrets.kdbx.testexec.kdbx`

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create KEEPASS storage: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no KEEPASS storage created")
		} else {

			_, err := sto.Get("missingkey")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty (error detected): %v", err)
			}

			key := "testcasePutGet"
			secret := "mysecret01"
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

func TestGroupKPPutGet(t *testing.T) {

	custom := map[string]any{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "tester01",
		"DefaultGroup":   "my_other_group",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create KEEPASS storage: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no KEEPASS storage created")
		} else {

			_, err := sto.Get("group01/missingkey")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty (error detected): %v", err)
			}

			key := "group02/testcasePutGet"
			secret := "mysecret05"
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

			key = "group99/testcasePut99Get"
			secret = "my99secret05"
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

func TestGroupNextKPPutGet(t *testing.T) {

	custom := map[string]any{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "tester01",
		"DefaultGroup":   "my_other_group",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create KEEPASS storage: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no KEEPASS storage created")
		} else {

			_, err := sto.Get("group02/missingkey")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty (error detected): %v", err)
			}

			key := "group02/testcasePutGet"
			secret := "mysecret05"
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

			key = "group02/testcasePut99Get"
			secret = "my99secret05"
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

func TestGroupEmptyKPPutGet(t *testing.T) {

	custom := map[string]any{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "tester01",
		"DefaultGroup":   "my_other_group",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create KEEPASS storage: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no KEEPASS storage created")
		} else {

			_, err := sto.Get("empty_missingkey")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty (error detected): %v", err)
			}

			key := "empty_testcasePutGet"
			secret := "mysecret05"
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

			key = "empty_testcasePut99Get"
			secret = "my99secret05"
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

func TestNewFileGet(t *testing.T) {

	keepassFile := fmt.Sprintf("../../tests/data/test_secrets_new.kdbx")

	custom := map[string]any{
		"Path":           keepassFile,
		"MasterPassword": "tester02",
		"DefaultGroup":   "my_new_group",
	}

	if _, err := os.Stat(keepassFile); err == nil {
		err = os.Remove(keepassFile)
		if err != nil {
			t.Errorf("failed to remove test file: %s\n%v", keepassFile, err)
		}
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create KEEPASS storage: %v", err)
	} else {
		if sto == nil {
			t.Errorf("no KEEPASS storage created")
		} else {

			_, err := sto.Get("group01/missingkey")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty (error detected): %v", err)
			}

			key := "group02/testcasePutGet"
			secret := "mysecret05"
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

			key = "group99/testcasePut99Get"
			secret = "my99secret05"
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
