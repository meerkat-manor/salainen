# Keypass secret storage

If you are saving your secrets in Keypass, then 
use this secret storage definition.

## Configuration

As Keepass requires a file and a master password, then these
need to be supplied in the configuration.  The master password
is handled as a **salainen** secret so you must supply in the
format "<provider>:<key>".  Some examples are:

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
