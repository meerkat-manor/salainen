# Using Bitwarden

If you are saving your secrets in Bitwarden, then 
use this secret storage definition.

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
format "<provider>:<key>".  Some examples are:

* plain:not_secure_password
* keyring:keepass_secret
