# Encrypted File secret storage

If you are using the encrypted file secret storage
then the contents are **encrypted** using the
default or specified key and the default or
specified encryption algorithm.

If no algorithm specified it defaults to 
ChaCha20-Poly1305

Sample configuration entry:

```json
        "efile": {
            "enabled": true,
            "name": "Encrypted File System",
            "custom": {
                "RootPath": "~/.secrets/salainen/efile",
                "Algorithm": "ChaCha20-Poly1305"
            }
        },
```

## SupportedEncryption algorithms

The following algorithms are supported:

* ChaCha20-Poly1305

## Golang Package

To set the secret value you call the function in your Go code as:

```go
salainen.Set("efile:<key>", "<value>")
```

This will use the default configuration and without 
a ``salainen.json`` or ``salainen.yml`` being in the current file 
search path or your home directory, it will enable **efile**
variables and files.

The prefix value **efile** indicates that this is an encrypted file provider secret.

If you call the register function with a configuration file location
then the sequence of calls is:

```go
salainen.Register("<config file>")
err := salainen.Set("efile:<key>", "<value>")
```

To fetch the secret, which could have been set outside of the
Go program, you get:

```go
salainen.Register("<config file>")
secret_value, err := salainen.Get("efile:<key>", "<value>")
```

## Command line

To set the secret outside of your program or script 
the use the "Set" function like so:

```
salainen 'efile:MY_SECRET' 'secretvalue'
```

You can retrieve the secret value like so:

```
salainen 'efile:MY_SECRET'
```

and the output is piped to the standard output, like terminal.
If there is an error then the program exit code is non zero (0)
