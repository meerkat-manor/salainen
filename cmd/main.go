package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/meerkat-manor/salainen"
	"github.com/meerkat-manor/salainen/config"
	"golang.design/x/clipboard"
)

func main() {

	configFile := flag.String("config", "", "path to config file")
	clip := flag.Bool("clip", false, "copy to clipboard the value fetched")
	help := flag.Bool("help", false, "help information")

	storage := flag.Bool("storage", false, "storage request")

	sync := flag.NewFlagSet("sync", flag.ExitOnError)
	fromStorage := sync.String("from", "", "from source storage")
	toStorage := sync.String("to", "", "to target storage")

	flag.Usage = PrintUsage

	flag.Parse()

	if *help {
		flag.Usage()

		// If there is config file then list storage types
		if *configFile != "" {
			PrintStorageTypes(configFile)
		}

		os.Exit(0)
		return
	}

	_, err := config.New(*configFile)
	if err != nil {
		log.Fatalf("processing aborted due to error: %v", err)
	}

	if *storage {
		err = process_storage(configFile, flag.Args())
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			os.Exit(1)
		} else {
			os.Exit(0)
		}
		return
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
				fmt.Printf("0.0.1\n")
				os.Exit(0)
				return
			}
		}
	}

	err = process_default(*clip, flag.Args())
	if err != nil {
		if !strings.HasPrefix(err.Error(), "wrong parameters") {
			fmt.Fprintf(os.Stderr, "An error occurred: %s\n", err)
		}
		os.Exit(1)
	} else {
		os.Exit(0)
	}

}

func process_default(clip bool, args []string) error {

	switch len(args) {
	case 1:
		val, err := salainen.Get(args[0])
		if err != nil {
			return err
		}

		if clip {
			err := clipboard.Init()
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
		err := salainen.Put(args[0], args[1])
		if err != nil {
			return err
		}
		return nil

	default:
		fmt.Fprintf(os.Stderr, "salainen [path] [value]\n")
		return fmt.Errorf("wrong parameters (%d)", len(args))
	}

}

func process_storage(configFile *string, args []string) error {

	if len(args) > 0 {
		match := false
		name := strings.ToLower(args[0])
		app, err := config.New(*configFile)

		if err == nil {
			for key, item := range app.StorageName {
				if strings.ToLower(key) == name {
					sstorage, err := salainen.GetSecretStorage(key)
					if err != nil {
						log.Fatalf("processing aborted due to error: %v", err)
					}
					fmt.Printf("Usage information for secret storage provider '%s' (%s) follows\n\n", name, item)
					sstorage.Help()
					match = true
					break
				}
			}
		}
		if !match {
			return fmt.Errorf("secret storage provider '%s' not recognised", name)
		}
		return nil

	} else {
		PrintStorageHelp(configFile)
		return nil
	}

}

func PrintUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s [flags] <secret key> <secret value (save)>\n", salainen.ProductName)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nMore command line information:\n")

	fmt.Fprintf(os.Stderr, "The secret key is mad up of two parts:\n")
	fmt.Fprintf(os.Stderr, "\t* storage type\n")
	fmt.Fprintf(os.Stderr, "\t* storage path (within the type)\n")
	fmt.Fprintf(os.Stderr, "The key takes the form of <type>:<path>, that is separated by colon (:)\n")

	fmt.Fprintf(os.Stderr, "If you only supply the <secret key> then this is a get secret action\n")
	fmt.Fprintf(os.Stderr, "If you supply the <secret key> and <secret value> then this is a set action of a secret\n\n")
	fmt.Fprintf(os.Stderr, "Using the -clip flag during secret get saves the value to the clipboard\n")
	fmt.Fprintf(os.Stderr, "You can provide a configuration file for setting storage attributes\n")
	fmt.Fprintf(os.Stderr, "Storage attributes are custom to each type\n")

	fmt.Fprintf(os.Stderr, "\nSample commands are:\n")
	fmt.Fprintf(os.Stderr, "\tsalainen wincred:db_password secret  --- saves the 'secret' to Windows credential under key db_password\n")
	fmt.Fprintf(os.Stderr, "\tsalainen wincred:db_password --- fetches the value from Windows credential under key db_password\n")
	fmt.Fprintf(os.Stderr, "\tsalainen keyring:db_password secret  --- saves the 'secret' to Linux keyring under key db_password\n")
	fmt.Fprintf(os.Stderr, "\tsalainen keyring:db_password --- fetches the value from Linux keyring under key db_password\n")
	fmt.Fprintf(os.Stderr, "\tsalainen efile:db_password.dat secret  --- saves the 'secret' to file named db_password.dat\n")
	fmt.Fprintf(os.Stderr, "\tsalainen --clip efile:db_password.dat  --- fetches the value to the clipboard from file named db_password.dat\n")

	fmt.Fprintf(os.Stderr, "\nDefining a configuration file allows control over type settings, such as Vault server\n")

	fmt.Fprintf(os.Stderr, "\nFor more information see https://github.com/meerkat-manor/salainen/cmd\n")
	fmt.Fprintf(os.Stderr, "\n(c) Copyright 2025 Merebox\n")
}

func PrintStorageTypes(configFile *string) {
	fmt.Fprintf(os.Stderr, "\nStorage types configured from: %s\n", *configFile)
	app, err := config.New(*configFile)
	if err != nil {
		log.Fatalf("processing aborted due to error: %v", err)
	}

	for key, item := range app.StorageName {
		fmt.Fprintf(os.Stderr, "\t%s (%s)\n", item, key)
	}

	fmt.Fprintf(os.Stderr, "\n(c) Copyright 2024 Merebox\n")
}

func PrintStorageHelp(configFile *string) {
	fmt.Fprintf(os.Stderr, "\nFor storage typeinformation help supply the name\n")
	fmt.Fprintf(os.Stderr, "\nStorage types configured from: %s\n", *configFile)
	app, err := config.New(*configFile)
	if err != nil {
		log.Fatalf("processing aborted due to error: %v", err)
	}

	for key, item := range app.StorageName {
		fmt.Fprintf(os.Stderr, "\t%s (%s)\n", item, key)
	}

	fmt.Fprintf(os.Stderr, "\n(c) Copyright 2024 Merebox\n")
}
