package main

import (
	"testing"

	"github.com/meerkat-manor/salainen/config"
)

func TestSimpleInit(t *testing.T) {

	configFile := "../defaults.json"
	appRun, err := config.New(configFile, false)
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

		arg01 := []string{
			"wincred:xyz",
		}
		err = process_default(false, arg01)
		if err != nil {
			t.Errorf("error with default process: %v", err)
		}
		arg02 := append(arg01, "miscrete")
		err = process_default(false, arg02)
		if err != nil {
			t.Errorf("error with default process: %v\n %v", arg02, err)
		}
		err = process_default(false, arg01)
		if err != nil {
			t.Errorf("error with default process: %v", err)
		}
	}

}

func TestStorage01(t *testing.T) {

	configFile := "../defaults.json"
	appRun, err := config.New(configFile, false)
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

		err = process_storage(&configFile, []string{})
		if err != nil {
			t.Errorf("error listing storage")
		}

		arg01 := []string{
			"rubbish",
		}
		err = process_storage(&configFile, arg01)
		if err == nil {
			t.Errorf("error not detected")
		}

		arg01[0] = "env"
		err = process_storage(&configFile, arg01)
		if err != nil {
			t.Errorf("error with ENV storage")
		}
	}

}
