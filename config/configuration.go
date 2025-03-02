package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/meerkat-manor/salainen"
	"github.com/meerkat-manor/salainen/extensions/bitwarden"
	"github.com/meerkat-manor/salainen/extensions/ejsons"
	"github.com/meerkat-manor/salainen/extensions/encryptedfile"
	"github.com/meerkat-manor/salainen/extensions/env"
	"github.com/meerkat-manor/salainen/extensions/file"
	"github.com/meerkat-manor/salainen/extensions/keepass"
	"github.com/meerkat-manor/salainen/extensions/keyring"
	"github.com/meerkat-manor/salainen/extensions/plain"
	"github.com/meerkat-manor/salainen/extensions/promptsec"
	"github.com/meerkat-manor/salainen/extensions/vault"
	"github.com/meerkat-manor/salainen/extensions/wincred"
	"gopkg.in/yaml.v2"
)

var debugState = false

type ProviderType string

const (
	Level1ProviderType ProviderType = "Level1"
	Level2ProviderType ProviderType = "Level2"
)

type StorageConfiguration struct {
	Enabled      bool         `yaml:"enabled" json:"enabled"`
	ProviderType ProviderType `yaml:"provider_type" json:"provider_type"`
	Name         string       `yaml:"name" json:"name"`
	Config       string       `yaml:"config" json:"config"`

	//	Custom interface{} `yaml:"custom" json:"custom"`
	Custom map[string]string `yaml:"custom" json:"custom"`
}

type ApplicationConfiguration struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`

	Storage map[string]StorageConfiguration `yaml:"providers" json:"providers"`
}

type ApplicationRun struct {
	Name    string
	Version string

	StorageName map[string]string
}

func New(configFile string, ignoreProviderErrors bool) (*ApplicationRun, error) {

	// Check for debug
	debugEnv := strings.ToUpper(os.Getenv(strings.ToUpper(salainen.ProductName) + "_DEBUG"))
	debugState = debugEnv == "TRUE" || debugEnv == "YES" || debugEnv == "1"

	// Check if config file exists
	if configFile != "" {

		if strings.HasPrefix(configFile, "~/") || strings.HasPrefix(configFile, "~\\") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			configFile = filepath.Join(homeDir, configFile[2:])
		}

		fi, err := os.Stat(configFile)
		if err != nil {
			return nil, fmt.Errorf("configuration file '%s' not found", configFile)
		}
		if fi.IsDir() {
			return nil, fmt.Errorf("configuration file '%s' is a directory", configFile)
		}

	} else {
		// Search order when config file not supplied
		// - environment SALAINEN value
		// - current path for product name (salainen) + .json or .yaml or .yml file
		// - path './config' for product name (salainen) + .json or .yaml or .yml file
		// - path './conf' for product name (salainen) + .json or .yaml or .yml file
		// - home directory ~/.salainen (product name) path for product name + .json or .yaml or .yml file
		// - home directory ~/.secrets path for product name + .json or .yaml or .yml file
		// - directory /etc for product name + .json or .yaml or .yml file

		// If a matching file name is found then it is assumed to contain valid configuration

		match := false
		// Search for config file
		tFile := ""

		if !match {
			tFile = os.Getenv(strings.ToUpper(salainen.ProductName))
			match = tFile != ""
		}

		if !match {
			tFile = checkConfigurationFile("")
			match = tFile != ""
		}

		if !match {
			tFile = checkConfigurationFile("./config")
			match = tFile != ""
		}

		if !match {
			tFile = checkConfigurationFile("./conf")
			match = tFile != ""
		}

		if !match {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				tFile = checkConfigurationFile(filepath.Join(homeDir, "."+salainen.ProductName))
				match = tFile != ""
			}
		}

		if !match {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				tFile = checkConfigurationFile(filepath.Join(homeDir, ".secrets"))
				match = tFile != ""
			}
		}

		if !match {
			tFile = checkConfigurationFile(filepath.Join("/etc", salainen.ProductName))
			match = tFile != ""
		}

		if match {
			configFile = tFile
		}
	}

	if debugState {
		if configFile == "" {
			fmt.Printf("DEBUG>> no configuration file set\n")
		} else {
			fmt.Printf("DEBUG>> configuration file set to: %s\n", configFile)
		}
	}

	app := ApplicationRun{
		Name:        salainen.ProductName,
		Version:     salainen.ProductVersion,
		StorageName: map[string]string{},
	}

	// Load configuration
	if configFile != "" {
		conf, err := loadConfig(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration")
		}

		app.Name = conf.Name
		app.Version = conf.Version

		pTypes := []ProviderType{
			Level1ProviderType,
			Level2ProviderType,
		}

		// Iterate over configuration in right sequence
		for _, itemPType := range pTypes {
			for key, item := range conf.Storage {

				// Default provider type
				if item.ProviderType == "" {
					item.ProviderType = Level1ProviderType
				}

				if item.Enabled && item.ProviderType == itemPType {

					switch key {

					case "plain":
						app.StorageName[key] = item.Name
						err := plain.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "env":
						app.StorageName[key] = item.Name
						err := env.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "wincred":
						app.StorageName[key] = item.Name
						err := wincred.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "keyring":
						app.StorageName[key] = item.Name
						err := keyring.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "file":
						app.StorageName[key] = item.Name
						err := file.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "efile":
						app.StorageName[key] = item.Name
						err := encryptedfile.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "ejson":
						app.StorageName[key] = item.Name
						err := ejsons.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "prompt":
						app.StorageName[key] = item.Name
						err := promptsec.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "bitwarden":
						app.StorageName[key] = item.Name
						err := bitwarden.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "keepass":
						app.StorageName[key] = item.Name
						err := keepass.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					case "vault":
						app.StorageName[key] = item.Name
						err := vault.Register(configFile, item.Custom)
						if err != nil && !ignoreProviderErrors {
							return nil, err
						}

					default:
						return nil, fmt.Errorf("provider '%s' not recognized", key)
					}
				}
			}
		}

	} else {
		// Load defaults

		custom := map[string]any{
			"Prefix": "{{.ProductName}}",
		}

		app.StorageName["plain"] = "Plain text"
		err := plain.Register("", custom)
		if err != nil {
			return nil, err
		}

		app.StorageName["env"] = "Environmental Variables"
		err = env.Register("", custom)
		if err != nil {
			return nil, err
		}

		app.StorageName["file"] = "File System"
		err = file.Register("", nil)
		if err != nil && !ignoreProviderErrors {
			return nil, err
		}

		app.StorageName["efile"] = "Encrypted File System"
		err = encryptedfile.Register("", nil)
		if err != nil && !ignoreProviderErrors {
			return nil, err
		}

		customKR := map[string]any{
			"Service": "{{.ProductName}}",
		}
		app.StorageName["keyring"] = "Keyring"
		err = keyring.Register("", customKR)
		if err != nil && !ignoreProviderErrors {
			return nil, err
		}

		app.StorageName["prompt"] = "Prompt"
		err = promptsec.Register("", nil)
		if err != nil && !ignoreProviderErrors {
			return nil, err
		}

		// If Windows
		if runtime.GOOS == "windows" {
			app.StorageName["wincred"] = "Windows Credential Manager"
			err := env.Register("", custom)
			if err != nil && !ignoreProviderErrors {
				return nil, err
			}
		}

		// If Linux
		if runtime.GOOS == "linux" {
			/* paused until implementation TODO
			 */
		}

	}

	return &app, nil
}

func loadConfig(configFile string) (*ApplicationConfiguration, error) {

	fi, err := os.Stat(configFile)
	if err != nil {
		return nil, fmt.Errorf("configuration file '%s' not found", configFile)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("configuration file '%s' is a directory", configFile)
	}

	var dataBuf []byte

	dataBuf, err = os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file '%s'. Error: %v", configFile, err)
	}

	appConfig := ApplicationConfiguration{}

	if strings.HasSuffix(strings.ToLower(configFile), ".json") {
		err := json.Unmarshal(dataBuf, &appConfig)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON configuration file '%s'. Error: %v", configFile, err)
		}

		// Under DEBUG Save a YAML copy, just to ensure we have up to date YAML version
		if debugState {
			content, errY := yaml.Marshal(appConfig)
			if errY == nil {
				oFileName := configFile + ".yaml"
				os.WriteFile(oFileName, content, os.ModeAppend)
			}
		}

	} else {

		err := yaml.Unmarshal(dataBuf, &appConfig)
		if err != nil {
			return nil, fmt.Errorf("error parsing YAML configuration file '%s'. Error: %v", configFile, err)
		}
	}

	return &appConfig, nil
}

func checkConfigurationFile(path string) string {

	if debugState {
		fmt.Printf("DEBUG>> testing configuration file path: %s\n", path)
	}

	tFile := filepath.Join(path, salainen.ProductName+".json")
	fi, err := os.Stat(tFile)
	if err == nil && !fi.IsDir() {
		return tFile
	}

	tFile = filepath.Join(path, salainen.ProductName+".yaml")
	fi, err = os.Stat(tFile)
	if err == nil && !fi.IsDir() {
		return tFile
	}

	tFile = filepath.Join(path, salainen.ProductName+".yml")
	fi, err = os.Stat(tFile)
	if err == nil && !fi.IsDir() {
		return tFile
	}

	return ""
}
