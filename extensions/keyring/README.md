# Keyring secret storage

Implements the system keyring, which is found as a capability
on Linux and macOS.

Sample configuration entry:

```json
        "keyring": {
            "enabled": true,
            "name": "Keyring",
            "custom": {
                "Service": "{{.ProductName}}"
            }
        },
```

## Golang Package

To set the secret value you call the function in your Go code as:

```go
salainen.Set("keyring:<key>", "<value>")
```

This will use the default configuration and without 
a ``salainen.json`` or ``salainen.yml`` being in the current file 
search path or your home directory, it will enable **keyring**
variables and files.

The prefix value **keyring** indicates that this is a keyring provider secret.

If you call the register function with a configuration file location
then the sequence of calls is:

```go
salainen.Register("<config file>")
err := salainen.Set("keyring:<key>", "<value>")
```

To fetch the secret, which could have been set outside of the
Go program, you get:

```go
salainen.Register("<config file>")
secret_value, err := salainen.Get("keyring:<key>", "<value>")
```

## Command line

To set the secret outside of your program or script 
the use the "Set" function like so:

```
salainen 'keyring:MY_SECRET' 'secretvalue'
```

You can retrieve the secret value like so:

```
salainen 'keyring:MY_SECRET'
```

and the output is piped to the standard output, like terminal.
If there is an error then the program exit code is non zero (0)

