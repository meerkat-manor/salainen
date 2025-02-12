# File secret storage

If you are using the plain file secret storage
then the contents are not encrypted.

The security of your secret is dependant on
securing the file.


Sample configuration entry:

```json
        "file": {
            "enabled": true,
            "name": "File System",
            "custom": {
                "RootPath": "~/.secrets/salainen/plainfile"
            }
        },
```


## Golang Package

To set the secret value you call the function in your Go code as:

```go
salainen.Set("file:<key>", "<value>")
```

This will use the default configuration and without 
a ``salainen.json`` or ``salainen.yml`` being in the current file 
search path or your home directory, it will enable **file**
variables and files.

The prefix value **file** indicates that this is a file provider secret.

If you call the register function with a configuration file location
then the sequence of calls is:

```go
salainen.Register("<config file>")
err := salainen.Set("file:<key>", "<value>")
```

To fetch the secret, which could have been set outside of the
Go program, you get:

```go
salainen.Register("<config file>")
secret_value, err := salainen.Get("file:<key>", "<value>")
```

## Command line

To set the secret outside of your program or script 
the use the "Set" function like so:

```
salainen 'file:MY_SECRET' 'secretvalue'
```

You can retrieve the secret value like so:

```
salainen 'file:MY_SECRET'
```

and the output is piped to the standard output, like terminal.
If there is an error then the program exit code is non zero (0)
