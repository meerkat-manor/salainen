package config

import (
	"testing"
)

func TestSimpleInit(t *testing.T) {

	configFile := "../defaults.json"

	appRun, err := New(configFile, false)
	if err != nil {
		t.Errorf("failed to create APPRUN")
	} else {
		if appRun == nil {
			t.Errorf("no APPRUN created")
		} else {
			t.Logf("success")
		}
	}

	if appRun != nil {
		if appRun.Version == "" {
			t.Errorf("no APPRUN version")
		}
		if appRun.Name == "" {
			t.Errorf("no APPRUN name")
		}

		sto, exists := appRun.StorageName["env"]
		if !exists {
			t.Errorf("failed to load ENV")
		} else {
			if sto == "" {
				t.Errorf("Missing ENV name")
			}
		}

	}

}
