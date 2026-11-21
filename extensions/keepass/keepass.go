package keepass

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/meerkat-manor/salainen"
	"github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

var debugState atomic.Bool

type f struct {
	ProviderName   string
	Path           string
	MasterPassword string
	DefaultGroup   string
	PrimaryGroup   string
	db             *gokeepasslib.Database
	mu             sync.Mutex
}

const providerName = "keepass"

func (sl *f) Init(custom interface{}) error {

	// Check for debug
	debugEnv := strings.ToUpper(os.Getenv(strings.ToUpper(salainen.ProductName) + "_DEBUG"))
	debugState.Store(debugEnv == "TRUE" || debugEnv == "YES" || debugEnv == "1")

	sl.ProviderName = "Keepass" // Note the leading capital
	sl.PrimaryGroup = salainen.ProductName + "_managed"

	if custom != nil {
		if settings, ok := custom.(map[string]string); ok {
			value, exists := settings["Path"]
			if exists && value != "" {
				sl.Path = value
				if sl.Path == "{{.ProductName}}" {
					sl.Path = salainen.ProductName
				}
			}
			value, exists = settings["MasterPassword"]
			if exists && value != "" {
				sl.MasterPassword = value
			}
			value, exists = settings["DefaultGroup"]
			if exists && value != "" {
				sl.DefaultGroup = value
			}
		} else {
			if debugState.Load() {
				fmt.Printf("DEBUG>> no custom configuration loaded\n")
			}
		}
	}

	if sl.DefaultGroup == "" {
		sl.DefaultGroup = salainen.ProductName
	}

	if sl.Path == "" {
		return fmt.Errorf("%s file location not specified in configuration under 'Path'", providerName)
	}

	if sl.MasterPassword == "" {
		return fmt.Errorf("%s master password not specified in configuration under 'MasterPassword'", providerName)
	}

	fpath, err := resolveKeepassPath(sl.Path)
	if err != nil {
		return err
	}

	if strings.HasPrefix(sl.MasterPassword, "keepass:") {
		return fmt.Errorf("error fetching %s access password with looping detected", sl.ProviderName)
	}

	password, errS := salainen.Get(sl.MasterPassword)
	if errS != nil || password == "" {
		return fmt.Errorf("error fetching %s access password.  More information: %v", sl.ProviderName, errS)
	}

	_, err = os.Stat(fpath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		rootGroup := gokeepasslib.NewGroup()
		rootGroup.Name = sl.PrimaryGroup

		sl.db = &gokeepasslib.Database{
			Header:      gokeepasslib.NewHeader(),
			Credentials: gokeepasslib.NewPasswordCredentials(password),
			Content: &gokeepasslib.DBContent{
				Meta: gokeepasslib.NewMetaData(),
				Root: &gokeepasslib.RootData{
					Groups: []gokeepasslib.Group{rootGroup},
				},
			},
		}
		return nil
	}

	dbFile, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer dbFile.Close()

	sl.db = gokeepasslib.NewDatabase()
	sl.db.Credentials = gokeepasslib.NewPasswordCredentials(password)
	if err := gokeepasslib.NewDecoder(dbFile).Decode(sl.db); err != nil {
		return err
	}

	return nil
}

func (sl *f) Put(path, value string) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	parts := strings.SplitN(path, "|", 2)
	npath := strings.SplitN(parts[0], "/", 2)

	kpath := npath[0]
	kgroup := sl.DefaultGroup
	if len(npath) > 1 {
		kgroup = npath[0]
		kpath = npath[1]
	}

	if sl.db == nil {
		return fmt.Errorf("internal error with Keepass database access for unlock")
	}

	if err := sl.db.UnlockProtectedEntries(); err != nil {
		return err
	}
	encodingStarted := false
	defer func() {
		if !encodingStarted {
			_ = sl.db.LockProtectedEntries()
		}
	}()

	if len(sl.db.Content.Root.Groups) < 1 {
		group := gokeepasslib.NewGroup()
		group.Name = sl.PrimaryGroup
		sl.db.Content.Root.Groups = append(sl.db.Content.Root.Groups, group)
	}

	gidx, eidx, _, _ := sl.get(kgroup, kpath)
	if gidx < 0 || eidx < 0 {
		entry := gokeepasslib.NewEntry()
		entry.Values = append(entry.Values, mkValue("Title", salainen.ProductName))
		entry.Values = append(entry.Values, mkValue("UserName", kpath))
		entry.Values = append(entry.Values, mkProtectedValue("Password", value))
		entry.Tags = salainen.ProductName + "_" + salainen.ProductVersion

		if gidx < 0 {
			group := gokeepasslib.NewGroup()
			group.Name = kgroup
			group.Entries = append(group.Entries, entry)
			sl.db.Content.Root.Groups[0].Groups = append(sl.db.Content.Root.Groups[0].Groups, group)
		} else {
			groupEntry := sl.db.Content.Root.Groups[0].Groups[gidx]
			groupEntry.Entries = append(groupEntry.Entries, entry)
			sl.db.Content.Root.Groups[0].Groups[gidx] = groupEntry
		}
	} else {
		matched := false
		for vidx, item := range sl.db.Content.Root.Groups[0].Groups[gidx].Entries[eidx].Values {
			if item.Key == "Password" {
				sl.db.Content.Root.Groups[0].Groups[gidx].Entries[eidx].Values[vidx] = mkProtectedValue("Password", value)
				if sl.db.Content.Root.Groups[0].Groups[gidx].Entries[eidx].Tags == "" {
					sl.db.Content.Root.Groups[0].Groups[gidx].Entries[eidx].Tags = salainen.ProductName + "_" + salainen.ProductVersion
				}
				matched = true
				break
			}
		}

		if !matched {
			return fmt.Errorf("unexpected processing error with existing entry")
		}
	}

	if err := sl.db.LockProtectedEntries(); err != nil {
		return err
	}
	encodingStarted = true

	fpath, err := resolveKeepassPath(sl.Path)
	if err != nil {
		return err
	}
	parent := filepath.Dir(fpath)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return err
	}

	file, err := os.CreateTemp(parent, ".salainen-keepass-*")
	if err != nil {
		return err
	}
	tmpName := file.Name()
	defer os.Remove(tmpName)
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return err
	}

	if err := gokeepasslib.NewEncoder(file).Encode(sl.db); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, fpath); err != nil {
		if removeErr := os.Remove(fpath); removeErr != nil && !os.IsNotExist(removeErr) {
			return err
		}
		if err := os.Rename(tmpName, fpath); err != nil {
			return err
		}
	}
	return os.Chmod(fpath, 0o600)
}

func (sl *f) get(kgroup string, kpath string) (int, int, string, error) {

	for gidx, group := range sl.db.Content.Root.Groups[0].Groups {

		if group.Name == kgroup {
			for eidx, entry := range sl.db.Content.Root.Groups[0].Groups[gidx].Entries {
				if entry.GetContent("UserName") == kpath {
					value := entry.GetPassword()
					if value == "" {
						return gidx, eidx, "", fmt.Errorf("failed to find value for: %s", kpath)
					}
					return gidx, eidx, value, nil
				}
			}
			return gidx, -1, "", fmt.Errorf("failed to find value for: %s", kpath)
		}
	}

	return -1, -1, "", fmt.Errorf("no matching key path was found")
}

func (sl *f) Get(path string) (string, error) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	parts := strings.SplitN(path, "|", 2)
	npath := strings.SplitN(parts[0], "/", 2)

	kpath := npath[0]
	kgroup := sl.DefaultGroup
	if len(npath) > 1 {
		kgroup = npath[0]
		kpath = npath[1]
	}

	if sl.db == nil {
		return "", fmt.Errorf("internal error with Keepass database access for unlock")
	}

	if errDb := sl.db.UnlockProtectedEntries(); errDb != nil {
		return "", errDb
	}
	defer func() {
		_ = sl.db.LockProtectedEntries()
	}()

	if len(sl.db.Content.Root.Groups) < 1 {
		return "", fmt.Errorf("failed to find value for (G): %s", kpath)
	}

	gidx, eidx, value, err := sl.get(kgroup, kpath)
	if err != nil {
		return "", fmt.Errorf("error with %s provider: %v", providerName, err)
	}
	if gidx < 0 || eidx < 0 {
		return "", fmt.Errorf("no matching key path was found")

	}
	if value == "" {
		return "", fmt.Errorf("failed to find value for: %s", kpath)
	}
	return value, nil

}

func (sl *f) Help() {
	fmt.Printf("Keepass help\n\n")
	fmt.Printf("Keepass can be used as a secret provider by using the\n")
	fmt.Printf("prefix 'keepass:' followed by the group and entry path.\n")
	fmt.Printf("\n")
	fmt.Printf("Only one file can be defined for the Keepass file, the\n")
	fmt.Printf("value of which is in the configuration under 'Path'.\n")
	fmt.Printf("The master password is in the configuration under 'MasterPassword'.\n")
	fmt.Printf("The master password is processed as a 'salainen` value so\n")
	fmt.Printf("define it using the format '<provider>:<key>' where you could\n")
	fmt.Printf("use for example 'plain:masterpassword' or 'keyring:secretkey \n")
	fmt.Printf("You cannot use 'keepass:password' as that would cause an infinite loop.\n")
	fmt.Printf("\n")
	fmt.Printf("For more information please see %s/extensions/keepass/ \n", salainen.SourceForgeURL)
}

func New(config string, custom interface{}) (salainen.SecretStorage, error) {

	storage := f{}
	err := storage.Init(custom)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

func Register(config string, custom interface{}) error {

	storage, err := New(config, custom)
	if err != nil {
		return err
	}
	salainen.AddSecretStorage(providerName, storage)

	return nil
}

func resolveKeepassPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}

func mkValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func mkProtectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: w.NewBoolWrapper(true)},
	}
}
