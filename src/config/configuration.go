package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"merebox.com/salainen"
	"merebox.com/salainen/bitwarden"
	"merebox.com/salainen/file"
	"merebox.com/salainen/promptsec"
	"merebox.com/salainen/wincred"
)

type StorageConfiguration struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Name    string `yaml:"name" json:"name"`
	Config  string `yaml:"config" json:"config"`

	Custom interface{} `yaml:"custom" json:"custom"`
}

type ApplicationConfiguration struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`

	Storage map[string]StorageConfiguration `yaml:"storage" json:"storage"`
}

type ApplicationRun struct {
	Name    string
	Version string

	StorageName map[string]string
}

func New(configFile string) (*ApplicationRun, error) {

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
		match := false
		// Search for config file
		tFile := ""

		if !match {
			tFile = filepath.Join("./", salainen.ProductName+".json")
			fi, err := os.Stat(tFile)
			if err == nil && !fi.IsDir() {
				match = true
			}
		}

		if !match {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				tFile = filepath.Join(homeDir, "."+salainen.ProductName, salainen.ProductName+".json")
				fi, err := os.Stat(tFile)
				if err == nil && !fi.IsDir() {
					match = true
				}
			}
		}

		if !match {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				tFile = filepath.Join(homeDir, ".secrets", salainen.ProductName+".json")
				fi, err := os.Stat(tFile)
				if err == nil && !fi.IsDir() {
					match = true
				}
			}
		}

		if !match {
			tFile = filepath.Join("/etc", salainen.ProductName, salainen.ProductName+".json")
			fi, err := os.Stat(tFile)
			if err == nil && !fi.IsDir() {
				match = true
			}
		}

		if match {
			configFile = tFile
		}
	}

	app := ApplicationRun{
		Name:        "salainen",
		Version:     "0.0.1",
		StorageName: map[string]string{},
	}

	// Load configuration
	if configFile != "" {
		conf, err := loadConfig(configFile)
		if err != nil {
			log.Fatalf("failed to load configuration")
		}

		app.Name = conf.Name
		app.Version = conf.Version

		for key, item := range conf.Storage {
			if item.Enabled {
				switch key {
				case "wincred":
					app.StorageName[key] = item.Name
					err := wincred.Register(configFile, item.Custom)
					if err != nil {
						return nil, err
					}

				case "file":
					app.StorageName[key] = item.Name
					err := file.Register(configFile, item.Custom)
					if err != nil {
						return nil, err
					}

				case "prompt":
					app.StorageName[key] = item.Name
					err := promptsec.Register(configFile, item.Custom)
					if err != nil {
						return nil, err
					}

				case "bitwarden":
					app.StorageName[key] = item.Name
					err := bitwarden.Register(configFile, item.Custom)
					if err != nil {
						return nil, err
					}

				}
			}

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
	} else {
		err := yaml.Unmarshal(dataBuf, appConfig)
		if err != nil {
			return nil, fmt.Errorf("error parsing YAML configuration file '%s'. Error: %v", configFile, err)
		}
	}

	return &appConfig, nil
}
