# Salainen

**Salinen** is a simplistic Go secrets manager whose objective is to provide
a consistent interface to retrieving secrets from various storage locations.

The storage locations supported are:

* Environment variables (*1)
* Files (*1)
* Windows credential manager (Windows)
* Bitwarden / Vaultwarden
* User prompt on terminal
* Text strings

(*1) Contents can be encrypted

Go programs can include the ``salainen`` package into code.

There is also a command line interface for **salainen**

## What is it?

**salainen** is a simple function to set and get secrets.

Configuration options exist for each location.  When encryption is
used, the algorithm and associated details ae defined in the 
configuration.

You can always stack your own encryption logic on the value
before calling **salainen** functions for that extra piece
of mind.

## How to use

First import package

```go
import "module merebox.com/salainen"
```

### Configuration

Configuration details are stored in a configuration file in either
JSON or YAML format.  Please use only the extensions ``.json'`,
``.yaml`` or ``.yml` for your configuration files.


The location is defined as an entry on a``map[string]interface{}``
and the map name is the storage location identifier.

The configuration settings are different for each location and are set in
the custom block.  The following settings are applicable to all locations and are 
the parent for the custom block.  

* enabled : A boolean value to indicate whether the location is available.
  If this item is not supplied and not set to ``true`` then the location is not
  available.
* name : The location English name.  If not supplied then the section name is used.
  This is only uuseful if listing available secret locations.
* custom : A custom definition for each locatio.  Please refer to the location 
  documentation for more details.  For example [file](./src/file/README.md)


An example configuration file is:

```json
{
    "name": "salainen",
    "version": "0.0.1",
    "storage": {
        "wincred": {
            "enabled": true,
            "name": "WinCred",
            "custom": {
                "Prefix": "{{.ProductName}}"
            }
        },
        "prompt": {
            "enabled": true,
            "name": "Prompt",
            "custom": {
                "Prefix": "{{.ProductName}}"
            }
        },
        "file": {
            "enabled": true,
            "name": "File System",
            "custom": {
                "RootPath": "~/.secrets/extras"
            }
        },
        "bitwarden": {
            "enabled": false,
            "name": "Bitwarden",
            "custom": {
                "ApiUrl": "",
                "IdentityUrl": "",
                "AccessToken" : ""
            }
        }
    }
}
```



### In Golang projects


#### Set

To set the secret value you call the function as

```go
salainen.Set("env:<key>")
```

This will use the default configuration and without 
a ``salainen.json`` or ``salainen.yml`` being in the current file 
search path or your home directory, it will enable **environmental**
variables and files.

The value **env** indicates that this is an environmental 
storge location secret.

If you call the register function with a configuration file location
then the sequence of calls is.  Once the locations are registered
you do not have to call the registration function within your 
current program function.

```go
salainen.Register("<config file")
salainen.Set("env:<key>")
```

### As command line

## Future storage locations

There are planned enhancements to support further
storage locations such as:

* Keyring (Linux)
* Keepass
* Git ( a variation on encrypted file )
* HashiCorp Vault
* Database (*1)
