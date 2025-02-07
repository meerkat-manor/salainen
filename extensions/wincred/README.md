# Windows credential secret storage

Uses the Microsoft Windows credential storage, which is found as a capability
on Microsoft Windows.

If you intend to use the software on multiple OS platforms, then the **keyring**
secret storage is a better alternative.

## Golang Package

To set the secret value you call the function in your Go code as:

```go
salainen.Set("wincred:<key>", "<value>")
```

This will use the default configuration and without 
a ``salainen.json`` or ``salainen.yml`` being in the current file 
search path or your home directory, it will enable **wincred**
variables and files.

_Note_: This storage method is only intended for Microsoft Windows.

The prefix value **wincred** indicates that this is a Windows Managed 
Credential storage location secret.

If you call the register function with a configuration file location
then the sequence of calls is:

```go
salainen.Register("<config file>")
err := salainen.Set("wincred:<key>", "<value>")
```

To fetch the secret, which could have been set outside of the
Go program, you get:

```go
salainen.Register("<config file>")
secret_value, err := salainen.Get("wincred:<key>", "<value>")
```

## Command line

To set the secret outside of your program or script 
the use the "Set" function like so:

```
salainen 'wincred:MY_SECRET' 'secretvalue'
```

You can retrieve the secret value like so:

```
salainen 'wincred:MY_SECRET'
```

and the output is piped to the standard output, like terminal.
If there is an error then the program exit code is non zero (0)

