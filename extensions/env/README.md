# Environment storage

Storing secrets in environment variables is common practice,
though its not considered secure in many use cases as the
value can be inspected.

## Golang Package

To set the secret value using the environment variable you call the function 
in your Go code as:

```go
salainen.Set("env:<key>", "<value>")
```

This will use the default configuration and without 
a ``salainen.json`` or ``salainen.yml`` being in the current file 
search path or your home directory, it will enable **environmental**
variables and files.

The prefix value **env** indicates that this is an environmental 
storage location secret.

If you call the register function with a configuration file location
then the sequence of calls is:

```go
salainen.Register("<config file>")
err := salainen.Set("env:<key>", "<value>")
```

To fetch the secret, which could have been set outside of the
Go program, you get:

```go
salainen.Register("<config file>")
secret_value, err := salainen.Get("env:<key>", "<value>")
```

## Command line

Using environment variables with command line is probably
overkill if all your secrets are stored that way, because
simply accessing them in you shell is simpler.

If you do use environment variables, then you should not
use the "Set" function, that is:

```
salainen 'env:MY_SECRET' 'secretvalue'
```

because the OS environment value applies only during
the salainen execution and the value is lost once the
child process exits.


## Windows environment

If you are using OS environment variables and
on Windows, then from a command line you need to
quote the key.

Your alternative request for a secret value stored in the 
Windows environment variable using the command line
is:

```
salainen '$env:MY_SECRET'
```

as an example for environment variable ```MY_SECRET``

## Linux


## Alternative

An alternative form exists for environment variables
by surrounding the variable name with `${` and `}`

```
salainen '${MY_SECRET}`
```

