package bitwarden

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/meerkat-manor/salainen"
)

const testFolderID = "123e4567-e89b-12d3-a456-426614174000"
const testItemID = "123e4567-e89b-12d3-a456-426614174001"

type bwPlainStorage struct {
	value string
	err   error
}

func (p *bwPlainStorage) Help() {}

func (p *bwPlainStorage) Init(custom interface{}) error {
	return nil
}

func (p *bwPlainStorage) Get(path string) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	return p.value, nil
}

func (p *bwPlainStorage) Put(path string, value string) error {
	return nil
}

func registerBWPlain(t *testing.T, value string, err error) {
	t.Helper()
	salainen.AddSecretStorage("plain", &bwPlainStorage{value: value, err: err})
	t.Cleanup(func() {
		delete(salainen.SecretStorages, "plain")
	})
}

type bwState struct {
	status         string
	unlockStatus   int
	folderStatus   int
	itemStatus     int
	passwordStatus int

	folderName string
	itemName   string
	username   string
	password   string

	folders      string
	items        string
	passwordBody string

	unlockBody string
}

func newBWState() *bwState {
	return &bwState{
		status:     "unlocked",
		folderName: "folder",
		itemName:   "item",
		username:   "user",
		password:   "secret",
	}
}

func writeBWResponse(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	if body != "" {
		_, _ = io.WriteString(w, body)
	}
}

func newBWServer(t *testing.T, state *bwState) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/status":
			writeBWResponse(w, http.StatusOK, fmt.Sprintf(
				`{"success":true,"data":{"object":"template","template":{"status":%q}}}`,
				state.status))

		case r.URL.Path == "/unlock":
			body, _ := io.ReadAll(r.Body)
			state.unlockBody = string(body)
			if state.unlockStatus != 0 {
				writeBWResponse(w, state.unlockStatus, "")
				return
			}
			writeBWResponse(w, http.StatusOK, `{"success":true}`)

		case r.URL.Path == "/list/object/folders":
			if state.folderStatus != 0 {
				writeBWResponse(w, state.folderStatus, "")
				return
			}
			if state.folders != "" {
				writeBWResponse(w, http.StatusOK, state.folders)
				return
			}
			writeBWResponse(w, http.StatusOK, fmt.Sprintf(
				`{"success":true,"data":{"object":"list","data":[{"object":"folder","id":%q,"name":%q}]}}`,
				testFolderID, state.folderName))

		case r.URL.Path == "/list/object/items":
			if state.itemStatus != 0 {
				writeBWResponse(w, state.itemStatus, "")
				return
			}
			if state.items != "" {
				writeBWResponse(w, http.StatusOK, state.items)
				return
			}
			writeBWResponse(w, http.StatusOK, fmt.Sprintf(
				`{"success":true,"data":{"object":"list","data":[{"name":%q,"login":{"username":%q,"password":%q}}]}}`,
				state.itemName, state.username, state.password))

		case strings.HasPrefix(r.URL.Path, "/object/password/"):
			if state.passwordStatus != 0 {
				writeBWResponse(w, state.passwordStatus, "")
				return
			}
			if state.passwordBody != "" {
				writeBWResponse(w, http.StatusOK, state.passwordBody)
				return
			}
			writeBWResponse(w, http.StatusOK, fmt.Sprintf(
				`{"success":true,"data":{"object":"password","data":%q}}`,
				state.password))

		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(server.Close)
	return server
}

func TestBitwardenNewAndInit(t *testing.T) {

	storage, err := New("", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	provider, ok := storage.(*f)
	if !ok {
		t.Fatalf("unexpected storage type")
	}
	if provider.ProductName != "Bitwarden" {
		t.Errorf("unexpected product name: %q", provider.ProductName)
	}

	storage, err = New("", map[string]string{
		"ApiUrl":      "http://api",
		"IdentityURL": "http://identity",
		"AccessToken": "plain:master",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	provider = storage.(*f)
	if provider.ApiUrl != "http://api" {
		t.Errorf("unexpected API URL: %q", provider.ApiUrl)
	}
	if provider.IdentityURL != "http://identity" {
		t.Errorf("unexpected identity URL: %q", provider.IdentityURL)
	}
	if provider.AccessToken != "plain:master" {
		t.Errorf("unexpected access token: %q", provider.AccessToken)
	}

	if _, err := New("", map[string]any{"ApiUrl": "wrong-type"}); err != nil {
		t.Errorf("unexpected error with non-string custom settings: %v", err)
	}
}

func TestBitwardenRegisterHelpPut(t *testing.T) {

	if err := Register("", nil); err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	t.Cleanup(func() {
		delete(salainen.SecretStorages, "bitwarden")
	})

	storage, err := salainen.GetSecretStorage("bitwarden")
	if err != nil {
		t.Fatalf("registered storage not found: %v", err)
	}

	storage.Help()

	if err := storage.Put("folder|item", "value"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBitwardenGetByUUID(t *testing.T) {

	state := newBWState()
	server := newBWServer(t, state)
	registerBWPlain(t, "master", nil)

	storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
	value, err := storage.Get(testItemID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "secret" {
		t.Errorf("unexpected value: %q", value)
	}
}

func TestBitwardenGetByFolder(t *testing.T) {

	state := newBWState()
	server := newBWServer(t, state)
	registerBWPlain(t, "master", nil)

	storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}

	value, err := storage.Get("folder/item")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "secret" {
		t.Errorf("unexpected value: %q", value)
	}

	value, err = storage.Get("folder/user")
	if err != nil {
		t.Fatalf("unexpected error matching username: %v", err)
	}
	if value != "secret" {
		t.Errorf("unexpected value: %q", value)
	}
}

func TestBitwardenGetByNestedFolder(t *testing.T) {

	state := newBWState()
	state.folderName = "parent/child"
	server := newBWServer(t, state)
	registerBWPlain(t, "master", nil)

	storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}

	value, err := storage.Get("parent/child/item")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "secret" {
		t.Errorf("unexpected value: %q", value)
	}
}

func TestBitwardenGetUnlocks(t *testing.T) {

	state := newBWState()
	state.status = "locked"
	server := newBWServer(t, state)
	registerBWPlain(t, "master", nil)

	storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
	if _, err := storage.Get(testItemID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(state.unlockBody, "master") {
		t.Errorf("unlock request did not include the master password: %q", state.unlockBody)
	}
}

func TestBitwardenGetSimplePath(t *testing.T) {

	state := newBWState()
	server := newBWServer(t, state)
	registerBWPlain(t, "master", nil)

	storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
	value, err := storage.Get("not-a-uuid-or-folder")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "" {
		t.Errorf("expected empty value, got %q", value)
	}
}

func TestBitwardenGetErrors(t *testing.T) {

	t.Run("no access token", func(t *testing.T) {
		storage := &f{}
		if _, err := storage.Get("path"); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("looping access token", func(t *testing.T) {
		storage := &f{AccessToken: "bitwarden:master"}
		if _, err := storage.Get("path"); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("access token lookup failure", func(t *testing.T) {
		registerBWPlain(t, "", fmt.Errorf("lookup failed"))
		storage := &f{AccessToken: "plain:master"}
		if _, err := storage.Get("path"); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("status request failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		url := server.URL
		server.Close()

		registerBWPlain(t, "master", nil)
		storage := &f{ApiUrl: url, AccessToken: "plain:master"}
		if _, err := storage.Get(testItemID); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("unlock failure", func(t *testing.T) {
		state := newBWState()
		state.status = "locked"
		state.unlockStatus = http.StatusUnauthorized
		server := newBWServer(t, state)

		registerBWPlain(t, "master", nil)
		storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
		if _, err := storage.Get(testItemID); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("folder request failure", func(t *testing.T) {
		state := newBWState()
		state.folderStatus = http.StatusInternalServerError
		server := newBWServer(t, state)

		registerBWPlain(t, "master", nil)
		storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
		if _, err := storage.Get("folder/item"); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("folder not found", func(t *testing.T) {
		state := newBWState()
		state.folders = `{"success":false,"data":{"object":"list","data":[]}}`
		server := newBWServer(t, state)

		registerBWPlain(t, "master", nil)
		storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
		if _, err := storage.Get("folder/item"); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("item request failure", func(t *testing.T) {
		state := newBWState()
		state.itemStatus = http.StatusInternalServerError
		server := newBWServer(t, state)

		registerBWPlain(t, "master", nil)
		storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
		if _, err := storage.Get("folder/item"); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("item not found", func(t *testing.T) {
		state := newBWState()
		state.itemName = "other"
		state.username = "other-user"
		server := newBWServer(t, state)

		registerBWPlain(t, "master", nil)
		storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
		if _, err := storage.Get("folder/item"); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("uuid request failure", func(t *testing.T) {
		state := newBWState()
		state.passwordStatus = http.StatusInternalServerError
		server := newBWServer(t, state)

		registerBWPlain(t, "master", nil)
		storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
		if _, err := storage.Get(testItemID); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("uuid unsuccessful response", func(t *testing.T) {
		state := newBWState()
		state.passwordBody = `{"success":false}`
		server := newBWServer(t, state)

		registerBWPlain(t, "master", nil)
		storage := &f{ApiUrl: server.URL, AccessToken: "plain:master"}
		if _, err := storage.Get(testItemID); err == nil {
			t.Errorf("expected error")
		}
	})
}
