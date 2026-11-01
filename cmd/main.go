package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/meerkat-manor/salainen"
	"github.com/meerkat-manor/salainen/config"
	"github.com/meerkat-manor/salainen/generate"
	"golang.design/x/clipboard"
)

const copyr = "(c) Copyright 2024-2026 Merebox"

func main() {

	configFile := flag.String("config", "", "path to config file")
	clip := flag.Bool("clip", false, "copy to clipboard the value fetched")
	overwrite := flag.Bool("overwrite", false, "overwrite existing value")
	genSecret := flag.Bool("generate", false, "generate a credential value and store it")
	help := flag.Bool("help", false, "help information")
	version := flag.Bool("version", false, fmt.Sprintf("%s version", salainen.ProductName))

	providers := flag.Bool("provider", false, "provider request")
	initProvider := flag.Bool("init", false, "initialise provider if not already initialised")

	sync := flag.NewFlagSet("sync", flag.ExitOnError)
	fromStorage := sync.String("from", "", "from source storage")
	toStorage := sync.String("to", "", "to target storage")

	flag.Usage = PrintUsage

	flag.Parse()

	// If there is no config supplied, then check for file
	// in default environment variable
	if configFile == nil || *configFile == "" {
		cf := os.Getenv(config.ENV_SALAINEN)
		if cf != "" {
			configFile = &cf
		}
	}

	if *help {
		flag.Usage()

		// If there is config file then list storage types
		if *configFile != "" {
			ProviderStorageTypes(configFile)
		}

		os.Exit(0)
		return
	}

	if *version {
		fmt.Printf("%s\n", salainen.ProductVersion)
		os.Exit(0)
		return
	}

	_, err := config.New(*configFile, *providers)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
		return
	}

	if *providers {
		err = process_storage(configFile, flag.Args(), *initProvider)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			os.Exit(1)
		} else {
			os.Exit(0)
		}
		return
	} else {
		if *initProvider {
			fmt.Fprint(os.Stderr, "-init command only applicable with -provider argument\n")
			os.Exit(0)
			return
		}
	}

	if len(flag.Args()) > 0 {
		switch flag.Arg(0) {
		case "sync":
			{
				sync.Parse(os.Args[2:])

				if fromStorage != nil {
					fmt.Printf("sync storage used: '%v' '%v'\n", *fromStorage, *toStorage)
				}

				os.Exit(0)
				return
			}
		case "version":
			{
				fmt.Printf("%s\n", salainen.ProductVersion)
				os.Exit(0)
				return
			}
		default:
			// Ignore and continue processing
		}
	}

	err = process_default(*clip, *genSecret, *overwrite, flag.Args())
	if err != nil {
		if !strings.HasPrefix(err.Error(), "wrong parameters") {
			fmt.Fprintf(os.Stderr, "An error occurred: %s\n", err)
		}
		os.Exit(1)
	} else {
		os.Exit(0)
	}

}

func process_default(clip bool, genSecret bool, overwrite bool, args []string) error {

	if genSecret && len(args) > 1 {
		return fmt.Errorf("you cannot supply a value to save when using generate option")
	}

	var err error
	var genCredential string

	if genSecret {
		genCredential, err = generate.AuthenticationSecret(nil)
		if err != nil {
			return err
		}
		args = append(args, genCredential)
	}

	switch len(args) {
	case 1:
		val, err := salainen.Get(args[0])
		if err != nil {
			return err
		}

		if clip {
			err = clipboard.Init()
			if err != nil {
				return err
			} else {
				clipboard.Write(clipboard.FmtText, []byte(val))
				fmt.Println("Secret copied too clipboard")
			}
		} else {
			fmt.Print(val)
		}
		return nil

	case 2:

		if !overwrite {
			// Check value does not exist
			_, err = salainen.Get(args[0])
			if err == nil {
				return fmt.Errorf("secret already exists and overwrite not specified. Save not completed.")
			}
		}

		err = salainen.Put(args[0], args[1])
		if err != nil {
			return err
		}

		if genSecret {
			if clip {
				err := clipboard.Init()
				if err != nil {
					return err
				} else {
					clipboard.Write(clipboard.FmtText, []byte(genCredential))
					fmt.Println("Generated secret copied too clipboard")
				}
			} else {
				fmt.Print(genCredential)
			}
		}

		return nil

	default:
		fmt.Fprintf(os.Stderr, "Please use argument 'salainen -help' for assistance\n")
		fmt.Fprintf(os.Stderr, "Common format is 'salainen provider:path [value]'\n")
		fmt.Fprintf(os.Stderr, "To list providers use 'salainen -provider'\n")
		return fmt.Errorf("wrong parameters (%d)", len(args))
	}

}

func process_storage(configFile *string, args []string, initProvider ...bool) error {

	doInit := false
	if len(initProvider) > 0 {
		doInit = initProvider[0]
	}

	filteredArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.EqualFold(arg, "-init") || strings.EqualFold(arg, "--init") {
			doInit = true
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}
	args = filteredArgs

	if len(args) > 0 {
		match := false
		name := strings.ToLower(args[0])
		app, err := config.New(*configFile, true)

		if err == nil {
			for key, item := range app.StorageName {
				if strings.ToLower(key) == name {
					sstorage, err := salainen.GetSecretStorage(key)
					if err != nil {
						return fmt.Errorf("processing aborted due to error: %v", err)
					}
					if doInit {
						return initialiseStorage(name, sstorage)
					}
					fmt.Printf("Usage information for secret storage provider '%s' (%s) follows\n\n", name, item)
					sstorage.Help()
					match = true
					break
				}
			}
		}
		if !match {
			return fmt.Errorf("secret storage provider '%s' not recognized or matched", name)
		}
		return nil

	} else {
		if doInit {
			return fmt.Errorf("a provider name must be supplied with -init")
		}
		PrintProviderHelp(configFile)
		return nil
	}

}

func initialiseStorage(name string, sstorage salainen.SecretStorage) error {

	initializer, ok := sstorage.(salainen.SecretStorageInitializer)
	if !ok {
		fmt.Printf("Provider '%s' does not require initialisation\n", name)
		return nil
	}

	initialised, err := initializer.Initialised()
	if err != nil {
		return fmt.Errorf("could not determine if provider '%s' is initialised: %v", name, err)
	}
	if initialised {
		fmt.Printf("Provider '%s' is already initialised\n", name)
		return nil
	}

	if err := initializer.Initialise(); err != nil {
		return fmt.Errorf("could not initialise provider '%s': %v", name, err)
	}

	fmt.Printf("Provider '%s' initialised\n", name)
	return nil
}

func PrintUsage() {
	w := os.Stderr
	fmt.Fprintf(w, "usage: %s [flags] <secret key> <secret value (save)>\n", salainen.ProductName)
	flag.PrintDefaults()
	fmt.Fprintf(w, "\nMore command line information:\n")

	fmt.Fprintf(w, "The secret key is made up of two parts:\n")
	fmt.Fprintf(w, "\t* provider type\n")
	fmt.Fprintf(w, "\t* storage key path (within the provider)\n")
	fmt.Fprintf(w, "The key takes the form of <type>%s<path>, that is separated by colon (%s)\n", salainen.ProviderSeparator, salainen.ProviderSeparator)

	fmt.Fprintf(w, "If you only supply the <secret key> then this is a get secret action\n")
	fmt.Fprintf(w, "If you supply the <secret key> and <secret value> then this is a set action of a secret\n\n")
	fmt.Fprintf(w, "Using the -clip flag during secret get saves the value to the clipboard\n")
	fmt.Fprintf(w, "You can provide a configuration file for setting provider attributes\n")
	fmt.Fprintf(w, "Provider attributes are custom to each type\n")
	fmt.Fprintf(w, "Use '-provider <type> -init' to initialise a provider if it is not already initialised\n")

	fmt.Fprintf(w, "\nSample commands are:\n")
	fmt.Fprintf(w, "\tsalainen wincred:db_password secret  --- saves the 'secret' to Windows credential under key db_password\n")
	fmt.Fprintf(w, "\tsalainen wincred:db_password --- fetches the value from Windows credential under key db_password\n")
	fmt.Fprintf(w, "\tsalainen keyring:db_password secret  --- saves the 'secret' to Linux keyring under key db_password\n")
	fmt.Fprintf(w, "\tsalainen keyring:db_password --- fetches the value from Linux keyring under key db_password\n")
	fmt.Fprintf(w, "\tsalainen efile:db_password.dat secret  --- saves the 'secret' to file named db_password.dat\n")
	fmt.Fprintf(w, "\tsalainen --clip efile:db_password.dat  --- fetches the value to the clipboard from file named db_password.dat\n")
	fmt.Fprintf(w, "\tsalainen -provider -init efile  --- initialises the efile provider, generating a GPG key when configured\n")

	fmt.Fprintf(w, "\nDefining a configuration file allows control over type settings, such as Vault server\n")

	fmt.Fprintf(w, "\nFor more information see %s/cmd\n", salainen.SourceForgeURL)
	fmt.Fprintf(w, "\n%s\n", copyr)
}

func ProviderStorageTypes(configFile *string) {
	w := os.Stderr
	fmt.Fprintf(w, "\nProvider types configured from: %s\n", *configFile)
	app, err := config.New(*configFile, true)
	if err != nil {
		log.Fatalf("processing aborted due to error: %v", err)
	}

	for key, item := range app.StorageName {
		fmt.Fprintf(w, "\t%s (%s)\n", item, key)
	}

	fmt.Fprintf(w, "\n%s\n", copyr)
}

func PrintProviderHelp(configFile *string) {

	w := os.Stderr
	fmt.Fprintf(w, "\nFor provider type information help include the name as argument\n")
	if configFile == nil || *configFile == "" {
		fmt.Fprintf(w, "\nProvider types are default values as no configuration file\n")
	} else {
		fmt.Fprintf(w, "\nProvider types configured from: %s\n", *configFile)
	}
	app, err := config.New(*configFile, true)
	if err != nil {
		log.Fatalf("processing aborted due to error: %v", err)
	}

	for key, item := range app.StorageName {
		fmt.Fprintf(w, "\t%s (%s)\n", item, key)
	}

	fmt.Fprintf(w, "\n%s\n", copyr)
}
