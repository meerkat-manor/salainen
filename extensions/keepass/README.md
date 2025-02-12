# Keypass secret storage

If you are saving your secrets in Keypass, then 
use this secret storage definition.

## Configuration

As Keepass requires a file and a master password, then these
need to be supplied in the configuration.  The master password
is handled as a **salainen** secret so you must supply in the
format "provider:key".  Some examples are:

* plain:not_secure_password
* keyring:keepass_secret

Sample configuration entry:

```json
        "keepass": {
            "enabled": true,
            "name": "Keepass",
            "custom": {
                "Path": "tests/data/test_secrets.kdbx",
                "MasterPassword" : "keyring:tester01",
                "DefaultGroup": "",
                "PrimaryGroup": ""
            }
        },
```

If the Keepass file does not exist it is created.

## Golang Package

To set the secret value you call the function in your Go code as:

```go
salainen.Set("keepass:<key>", "<value>")
```

This will use the default configuration and without 
a ``salainen.json`` or ``salainen.yml`` being in the current file 
search path or your home directory, it will enable **keepass**
variables and files.

The prefix value **keepass** indicates that this is a 
Keepass storage location secret.

If you call the register function with a configuration file location
then the sequence of calls is:

```go
salainen.Register("<config file>")
err := salainen.Set("keepass:<key>", "<value>")
```

To fetch the secret, which could have been set outside of the
Go program, you get:

```go
salainen.Register("<config file>")
secret_value, err := salainen.Get("keepass:<key>", "<value>")
```

## Command line

To set the secret outside of your program or script 
the use the "Set" function like so:

```
salainen 'keepass:MY_SECRET' 'secretvalue'
```

You can retrieve the secret value like so:

```
salainen 'keepass:MY_SECRET'
```

and the output is piped to the standard output, like terminal.
If there is an error then the program exit code is non zero (0)
