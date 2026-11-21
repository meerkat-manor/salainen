package ejsons

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Shopify/ejson"
	"github.com/meerkat-manor/salainen/extensions/plain"
)

/*

Test PRIVATE key:
 a423a03d8fc7794f9e3cbba98f5abb2528a6fc84663a110f3b15fd8642677391
Test PUBLIC key:
 bedd3c5bb7831fcadc82d83303f5da1bd8a36baf4fb91c85740ab80cfd7e770f

*/

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

func getTestDirectory() string {

	return "../../tests/ejson"
}

func TestEjsonRejectsPathTraversal(t *testing.T) {
	testInit()
	root := t.TempDir()
	storage, err := New("", map[string]string{
		"RootPath":    root,
		"PublicKey":   "68bbde4475f044afdb8869977ab68d17d6354acdec89908bfeeb0c4738803a15",
		"PrivateKey":  "plain:8fb61adf1cce80d2e205db976a6f99731e9c413ed0887b35b026fc11be377440",
		"ElementName": "password",
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	if err := storage.Put("../outside.json", "value"); err == nil {
		t.Errorf("expected traversal put to be rejected")
	}
	if _, err := storage.Get("../outside.json"); err == nil {
		t.Errorf("expected traversal get to be rejected")
	}
}

func TestEjsonRestoresConfiguredPublicKey(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not meaningful on Windows")
	}
	testInit()
	root := t.TempDir()
	const publicKey = "68bbde4475f044afdb8869977ab68d17d6354acdec89908bfeeb0c4738803a15"
	storage, err := New("", map[string]string{
		"RootPath":    root,
		"PublicKey":   publicKey,
		"PrivateKey":  "plain:8fb61adf1cce80d2e205db976a6f99731e9c413ed0887b35b026fc11be377440",
		"ElementName": "password",
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	file := filepath.Join(root, "secret.json")
	if err := storage.Put("secret.json", "value"); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read ejson file: %v", err)
	}
	var document map[string]interface{}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("failed to parse ejson file: %v", err)
	}
	document["_public_key"] = "attacker-controlled-key"
	tampered, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("failed to marshal tampered document: %v", err)
	}
	if err := os.WriteFile(file, tampered, 0o600); err != nil {
		t.Fatalf("failed to write tampered document: %v", err)
	}

	if err := storage.Put("secret.json", "updated"); err != nil {
		t.Fatalf("unexpected second put error: %v", err)
	}
	data, err = os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to reread ejson file: %v", err)
	}
	document = nil
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("failed to parse updated ejson file: %v", err)
	}
	if document["_public_key"] != publicKey {
		t.Errorf("configured public key was not restored: %v", document["_public_key"])
	}
}

func TestSimpleEjsonCreate(t *testing.T) {

	testInit()
	custom := map[string]string{
		"RootPath":    getTestDirectory(),
		"KeyDir":      "",
		"PublicKey":   "68bbde4475f044afdb8869977ab68d17d6354acdec89908bfeeb0c4738803a15",
		"PrivateKey":  "plain:8fb61adf1cce80d2e205db976a6f99731e9c413ed0887b35b026fc11be377440",
		"ElementName": "password",
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

func TestSimpleEjsonGet(t *testing.T) {

	testInit()
	custom := map[string]string{
		"RootPath":    getTestDirectory(),
		"KeyDir":      "",
		"PublicKey":   "",
		"PrivateKey":  "plain:8fb61adf1cce80d2e205db976a6f99731e9c413ed0887b35b026fc11be377440",
		"ElementName": "password",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create file storage")
	} else {
		if sto == nil {
			t.Errorf("no file storage created")
		} else {

			key := "testcaseEjsonPutGet_04_keep.json"
			secret := "mysecret01"

			value, err := sto.Get(key)
			if err != nil || value != secret {
				t.Errorf("%s :: %v", value, err)
			} else {
				t.Logf("success : %s", value)
			}

		}
	}

}

func TestSimpleEjsonPutGet(t *testing.T) {

	testInit()
	custom := map[string]string{
		"RootPath":    getTestDirectory(),
		"KeyDir":      "",
		"PublicKey":   "68bbde4475f044afdb8869977ab68d17d6354acdec89908bfeeb0c4738803a15",
		"PrivateKey":  "plain:8fb61adf1cce80d2e205db976a6f99731e9c413ed0887b35b026fc11be377440",
		"ElementName": "password",
	}

	sto, err := New("", custom)
	if err != nil {
		t.Errorf("failed to create file storage")
	} else {
		if sto == nil {
			t.Errorf("no file storage created")
		} else {

			value, err := sto.Get("missingkeyfile.json")
			if err == nil {
				t.Errorf("failed to flag error")
			} else {
				t.Logf("success with empty: %s", value)
			}

			key := "testcaseEjsonPutGet_03.json"
			secret := "mysecret01"

			// Remove prior file
			os.Remove(path.Join(custom["RootPath"], key))

			err = sto.Put(key, secret)
			if err != nil {
				t.Errorf("failed to put value: %v", err)
			} else {
				value, err := sto.Get(key)
				if err != nil || value != secret {
					t.Errorf("%s :: %v", value, err)
				} else {
					t.Logf("success : %s", value)
				}
			}

		}
	}

}

func TestCheckEjsonGeneration(t *testing.T) {

	pub, priv, err := ejson.GenerateKeypair()
	if err != nil {
		t.Errorf("failed to create file storage")
	} else {
		t.Logf("Test PRIVATE key:\n %s\n", priv)
		t.Logf("Test PUBLIC key:\n %s\n", pub)
	}

	fpath := path.Join(getTestDirectory(), "TestEjsonGeneration_01_keep.json")

	input, err := os.ReadFile(fpath)
	if err != nil {
		t.Error(err)
	}

	fpath = path.Join(getTestDirectory(), "TestEjsonGeneration_99.json")

	// Change pub key
	data := strings.Replace(string(input), "PUBKEY_CHANGE", pub, 1)

	err = os.WriteFile(fpath, []byte(data), 0o600)
	if err != nil {
		t.Error(err)
	}

	count, err := ejson.EncryptFileInPlace(fpath)
	if err != nil || count < 1 {
		t.Errorf("failed to encrypt file")
	}

	contents, err := ejson.DecryptFile(fpath, "", priv)
	if err != nil || count < 1 {
		t.Errorf("failed to decrypt file: %v", err)
	} else {
		t.Logf("contents: %s", string(contents))
	}
}
