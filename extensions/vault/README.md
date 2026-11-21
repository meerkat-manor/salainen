# Vault secret storage

If you are saving your secrets in HashiCorp Vault
or equivalent, then use this secret storage definition.

## Configuration

As Vault requires a URL and a master password, then these
need to be supplied in the configuration.  The master password
is handled as a **salainen** secret so you must supply in the
format "provider:key".  Some examples are:

* plain:not_secure_password
* keyring:keepass_secret

Sample configuration entry:

```json
        "vault": {
            "enabled": true,
            "name": "Vault",
            "custom": {
                "Path": "https://",
                "MasterPassword" : "keyring:tester01",
                "DefaultGroup": "",
                "PrimaryGroup": ""
            }
        },
```

If the Vault does not exist the request for setting or
getting a value will fail.

## Golang Package

To set the secret value you call the function in your Go code as:

```go
salainen.Set("vault:<key>", "<value>")
```

You will need to enable the **vault** provider
in you configuration file as it is not enabled by default.

The prefix value **vault** indicates that this is a 
Vault provider secret.

If you call the register function with a configuration file location
then the sequence of calls is:

```go
salainen.Register("<config file>")
err := salainen.Set("vault:<key>", "<value>")
```

To fetch the secret, which could have been set outside of the
Go program, you get:

```go
salainen.Register("<config file>")
secret_value, err := salainen.Get("vault:<key>", "<value>")
```

## Command line

When using the command line you will need to set the
configuration file, or the configuration file must
be in a defined search path.

```
salainen 'vault:MY_SECRET' 'secretvalue'
```

You can retrieve the secret value like so:

```
salainen 'vault:MY_SECRET'
```

and the output is piped to the standard output, like terminal.
If there is an error then the program exit code is non zero (0)
