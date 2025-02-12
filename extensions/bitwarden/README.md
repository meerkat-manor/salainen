# Using Bitwarden

If you are saving your secrets in Bitwarden, then 
use this secret storage definition.

You will need to have access to the Bitwarden local
server, not the remote hosted version.  To do this you need
to download the "bw.exe" and execute it locally

```bash
bw serve
```

## Configuration

As the Bitwarden service is accessed via an API
the API configuration details are required.  This includes:

* URL
* API Token

Sample configuration entry:

```json
        "bitwarden": {
            "enabled": true,
            "name": "Bitwarden",
            "custom": {
                "ApiUrl": "http://localhost:8087",
                "IdentityURL": "",
                "AccessToken" : "keyring:test_BW"
            }
        }
```


## Securing the Credential or API key

As Bitwarden requires secure API token itself this needs
to be stored locally.  One option is to store them
in a keyring or environment variable.

The keyring or environment variable can be
accessed using **salainen** so you must supply in the
format "provider:key".  Some examples are:

* env:variable_name
* keyring:bitwarden_secret



## Golang Package

To set the secret value you call the function in your Go code as:

```go
salainen.Set("bitwarden:<key>", "<value>")
```

You will need to enable the **bitwarden** provider
in you configuration file as it is not enabled by default.

The prefix value **bitwarden** indicates that this is a 
Bitwarden provider secret.

If you call the register function with a configuration file location
then the sequence of calls is:

```go
salainen.Register("<config file>")
err := salainen.Set("bitwarden:<key>", "<value>")
```

To fetch the secret, which could have been set outside of the
Go program, you get:

```go
salainen.Register("<config file>")
secret_value, err := salainen.Get("bitwarden:<key>", "<value>")
```

## Command line

To set the secret outside of your program or script 
the use the "Set" function like so:

```
salainen 'bitwarden:MY_SECRET' 'secretvalue'
```

You can retrieve the secret value like so:

```
salainen 'bitwarden:MY_SECRET'
```

and the output is piped to the standard output, like terminal.
If there is an error then the program exit code is non zero (0)

