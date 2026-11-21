package keepass

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

func TestSimpleKPCreate(t *testing.T) {

	testInit()
	custom := map[string]string{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "plain:tester01",
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

	testInit()
	custom := map[string]string{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "plain:tester01",
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

	testInit()
	custom := map[string]string{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "plain:tester01",
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

	testInit()
	custom := map[string]string{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "plain:tester01",
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

	testInit()
	custom := map[string]string{
		"Path":           fmt.Sprintf("../../tests/data/test_secrets.kdbx"),
		"MasterPassword": "plain:tester01",
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

func TestKeepassPasswordIsProtectedAndFileIsPrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not meaningful on Windows")
	}
	testInit()
	keepassFile := filepath.Join(t.TempDir(), "secrets.kdbx")
	storage, err := New("", map[string]string{
		"Path":           keepassFile,
		"MasterPassword": "plain:tester03",
	})
	if err != nil {
		t.Fatalf("failed to create keepass storage: %v", err)
	}
	provider, ok := storage.(*f)
	if !ok {
		t.Fatalf("unexpected storage type")
	}

	if err := storage.Put("protected-key", "secret-value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	found := false
	for _, group := range provider.db.Content.Root.Groups[0].Groups {
		for _, entry := range group.Entries {
			for _, item := range entry.Values {
				if item.Key == "Password" && item.Value.Protected.Bool {
					found = true
				}
			}
		}
	}
	if !found {
		t.Errorf("password value is not marked protected")
	}

	info, err := os.Stat(keepassFile)
	if err != nil {
		t.Fatalf("failed to stat keepass file: %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Errorf("keepass file has permissive mode %v", info.Mode().Perm())
	}
}

func TestNewFileGet(t *testing.T) {

	testInit()
	keepassFile := fmt.Sprintf("../../tests/data/test_secrets_new.kdbx")

	custom := map[string]string{
		"Path":           keepassFile,
		"MasterPassword": "plain:tester02",
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
