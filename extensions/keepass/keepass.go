package keepass

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/meerkat-manor/salainen"
	"github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

type f struct {
	Path         string
	Credential   string
	DefaultGroup string
	PrimaryGroup string
	db           *gokeepasslib.Database
}

const providerName = "keepass"

func (sl *f) Init(custom interface{}) error {

	sl.PrimaryGroup = salainen.ProductName + "_managed"

	if custom != nil {
		settings := custom.(map[string]interface{})
		value, exists := settings["Path"]
		if exists && value.(string) != "" {
			sl.Path = value.(string)
			if sl.Path == "{{.ProductName}}" {
				sl.Path = salainen.ProductName
			}
		}
		value, exists = settings["Credential"]
		if exists && value.(string) != "" {
			sl.Credential = value.(string)
		}
		value, exists = settings["DefaultGroup"]
		if exists && value.(string) != "" {
			sl.DefaultGroup = value.(string)
		}
	}

	if sl.DefaultGroup == "" {
		sl.DefaultGroup = salainen.ProductName
	}

	fpath := sl.Path
	if strings.HasPrefix(fpath, "~/") || strings.HasPrefix(fpath, "~\\") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		fpath = filepath.Join(homeDir, fpath[2:])
	}

	// Check file exists
	_, err := os.Stat(fpath)
	if err != nil {
		return err
	}

	dbFile, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer dbFile.Close()

	sl.db = gokeepasslib.NewDatabase()
	sl.db.Credentials = gokeepasslib.NewPasswordCredentials(sl.Credential)
	_ = gokeepasslib.NewDecoder(dbFile).Decode(sl.db)

	return nil
}

func (sl *f) Put(path, value string) error {
	parts := strings.SplitN(path, "|", 2)
	npath := strings.SplitN(parts[0], "/", 2)

	kpath := npath[0]
	kgroup := sl.DefaultGroup
	if len(npath) > 1 {
		kgroup = npath[0]
		kpath = npath[1]
	}

	sl.db.UnlockProtectedEntries()

	// Check if we have a group
	if len(sl.db.Content.Root.Groups) < 1 {
		// Add base group
		group := gokeepasslib.NewGroup()
		group.Name = sl.PrimaryGroup
		sl.db.Content.Root.Groups = append(sl.db.Content.Root.Groups, group)
	}

	// See if entry already exists
	gidx, eidx, _, _ := sl.get(kgroup, kpath)
	if gidx < 0 || eidx < 0 {
		entry := gokeepasslib.NewEntry()
		entry.Values = append(entry.Values, mkValue("Title", salainen.ProductName)) // TODO
		entry.Values = append(entry.Values, mkValue("UserName", kpath))
		entry.Values = append(entry.Values, mkValue("Password", value))
		entry.Tags = salainen.ProductName + "_" + salainen.ProductVersion

		if gidx < 0 {
			// Add the group
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
		// Find the entry
		matched := false
		for vidx, item := range sl.db.Content.Root.Groups[0].Groups[gidx].Entries[eidx].Values {
			if item.Key == "Password" {
				sl.db.Content.Root.Groups[0].Groups[gidx].Entries[eidx].Values[vidx] = mkValue("Password", value)
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

	sl.db.LockProtectedEntries()

	fpath := sl.Path
	if strings.HasPrefix(fpath, "~/") || strings.HasPrefix(fpath, "~\\") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		fpath = filepath.Join(homeDir, fpath[2:])
	}

	if sl.db == nil {
		return fmt.Errorf("keepass database not initialized")
	}

	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	keepassEncoder := gokeepasslib.NewEncoder(file)
	if err := keepassEncoder.Encode(sl.db); err != nil {
		fmt.Printf("failed to encode for: %s", kpath)
		return err
	}

	return nil
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
	parts := strings.SplitN(path, "|", 2)
	npath := strings.SplitN(parts[0], "/", 2)

	kpath := npath[0]
	kgroup := sl.DefaultGroup
	if len(npath) > 1 {
		kgroup = npath[0]
		kpath = npath[1]
	}

	sl.db.UnlockProtectedEntries()

	if len(sl.db.Content.Root.Groups) < 1 {
		return "", fmt.Errorf("failed to find value for (G): %s", kpath)
	}

	gidx, eidx, value, err := sl.get(kgroup, kpath)
	if err != nil {
		return "", err
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
	fmt.Printf("The master password is in the configuration under 'Credential'.\n")
	fmt.Printf("The master password is processed as a 'salainen` value so.\n")
	fmt.Printf("define it using the format '<provider>:<key>' where you could\n")
	fmt.Printf("use for example 'plain:masterpassword' or 'keyring:secretkey \n")
	fmt.Printf("You cannot use 'keepass:password' as that would cause an infinite loop.\n")
	fmt.Printf("\n")
	fmt.Printf("For more information please see %s/extensions/keepass/README.md \n", salainen.SourceForgeURL)
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

func mkValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func mkProtectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: w.NewBoolWrapper(true)},
	}
}
