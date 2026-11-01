# EJSON Git storage file

The **ejson** provider stores secrets in a JSON file. In the default
mode it uses the [Shopify ejson](https://github.com/Shopify/ejson)
format: the attribute names are readable and the values are encrypted.
This makes the file suitable for storing in Git.

An alternative `GPG` algorithm is also supported. In GPG mode the whole
JSON file is encrypted with a GPG public key, so both names and values
are hidden. Use the default EJSON mode when you need readable attribute
names in the repository.

## Supported algorithms

| Algorithm | Description |
| --- | --- |
| `EJSON` (default) | Shopify ejson format. Values are encrypted, names are readable. |
| `GPG` / `OpenPGP` | The complete JSON file is encrypted with GPG. |

## Summary of ejson format

Secrets are collected in a JSON file. In the default EJSON mode all the
string values are encrypted, public keys are embedded in the file, and
the decrypter looks up the corresponding private key from its local
filesystem.

_Note_: YAML formatted secret files are not supported.

## Configuration

| Setting | Description |
| --- | --- |
| `RootPath` | Directory where the JSON files are stored. |
| `Algorithm` | `EJSON` (default) or `GPG`. |
| `KeyDir` | Directory containing private keys (EJSON mode). |
| `PublicKey` | Public key used to encrypt values. |
| `PrivateKey` | Private key, optionally referenced from another provider. |
| `ElementName` | JSON element that holds the secret. Defaults to `password`. |
| `GPGExecutable` | GPG executable. Defaults to `gpg` from `PATH`. |
| `GPGHomeDir` | Optional GPG home directory. Useful for a dedicated keyring. |
| `GPGRecipient` | Key ID, fingerprint or email address used for encryption. |
| `GPGPassphrase` | Optional private key passphrase used for decryption. |
| `GPGMarker` | Optional path to the provider marker file created by `-init`. |
| `GPGName` | Name used when `-init` generates a key. |
| `GPGEmail` | Email used when `-init` generates a key. |
| `GPGExpireDate` | Key expiry used by `-init` (`0` means never). |
| `GPGKeyType` | Key type used by `-init`. Defaults to `RSA`. |
| `GPGKeyLength` | Key length used by `-init`. Defaults to `4096`. |

### EJSON example

```json
        "ejson": {
            "enabled": true,
            "name": "Encrypted JSON file",
            "custom": {
                "RootPath": "~/.secrets/salainen/",
                "KeyDir": "~/.secrets/salainen/",
                "PublicKey": "",
                "PrivateKey": "",
                "ElementName": "password",
                "Algorithm": "EJSON"
            }
        },
```

_Warning_: If the private key is supplied in the configuration and not through
the Key directory, then the value needs to be secured by another type of
**salainen** provider to ensure secrecy and not expose the key in configuration.
You could use providers such as `keyring:` and `wincred:` and of course
`plain:`. A `plain:` provider should only be used for testing **salainen**,
not for other situations.

### GPG example

```json
        "ejson": {
            "enabled": true,
            "name": "Encrypted JSON file (GPG)",
            "custom": {
                "RootPath": "~/.secrets/salainen/",
                "ElementName": "password",
                "Algorithm": "GPG",
                "GPGHomeDir": "~/.secrets/salainen/gnupg",
                "GPGRecipient": "me@example.com"
            }
        },
```

_Note_: `GPGRecipient` can be omitted after `-init` has generated a key.
The recipient is then read from the marker file in `RootPath`
(`.salainen-ejson-gpg.json`).

## Setup

### Install GPG

#### GPG on Linux

For Debian/Ubuntu:

```bash
sudo apt update
sudo apt install gnupg
```

For Fedora/RHEL:

```bash
sudo dnf install gnupg2
```

Verify the installation:

```bash
gpg --version
```

#### Windows

Install Gpg4win from [https://www.gpg4win.org/](https://www.gpg4win.org/)
or with `winget`:

```powershell
winget install GnuPG.Gpg4win
```

Alternatively install the standalone GnuPG package:

```powershell
winget install GnuPG.GnuPG
```

Open a new PowerShell window and verify the installation:

```powershell
gpg --version
```

If `gpg` is not found, add the GnuPG `bin` directory to `PATH`. For
Gpg4win this is normally:

```text
C:\Program Files (x86)\GnuPG\bin
```

### Generate a GPG key automatically

The provider can create the key pair for you. The command checks whether
the provider is already initialised. If it is not initialised, you are
prompted for the key name, email address, passphrase and expiry.

```text
salainen -provider -init ejson
```

The `-init` option may also be placed after the provider name:

```text
salainen -provider ejson -init
```

The generated key is written to the GPG keyring (`GPGHomeDir` when set)
and a marker file is written to `RootPath`. Running `-init` again reports
that the provider is already initialised.

For non-interactive setup, supply the values in the provider `custom`
configuration:

```json
            "custom": {
                "RootPath": "~/.secrets/salainen/",
                "Algorithm": "GPG",
                "GPGHomeDir": "~/.secrets/salainen/gnupg",
                "GPGName": "Salainen",
                "GPGEmail": "salainen@example.com",
                "GPGPassphrase": "",
                "GPGExpireDate": "0",
                "GPGKeyLength": "4096"
            }
```

### Generate a GPG key manually

You can use an existing key or create one with GPG directly.

#### Setup Linux

```bash
gpg --full-generate-key
```

Choose `RSA and RSA`, a key size of `4096`, and an expiry that suits
your environment. Enter your name, email address and a passphrase.

List the key and note the fingerprint:

```bash
gpg --list-secret-keys --keyid-format=long
```

#### Windows (PowerShell)

```powershell
gpg --full-generate-key
```

Use the same prompts as on Linux. Then list the key:

```powershell
gpg --list-secret-keys --keyid-format=long
```

Copy the fingerprint from the `sec` entry and set it as
`GPGRecipient`, or use the email address associated with the key.

## Usage

Store and version the file in Git or another location as required, but
ensure secrets are encrypted before storage. When you require a value,
call the **Get** function passing the file as the key and the private key
or GPG settings in the custom configuration for **ejson**.

Only one element can be decrypted and this is passed as the
**ElementName** in the custom configuration.

### Scenario with configuration file

1. Create a configuration file and include an element named
   `_public_key` at the root (EJSON mode only).
2. This configuration file can be saved in Git with the project.
   This could be a default or template configuration file.
3. Include all the other element names with a leading `_` symbol
   on the name. The prefix `_` indicates to the ejson processing
   that the value is not to be encrypted / decrypted.
4. Include the secret, say an element named `password`. This value
   is then encrypted and committed to Git.
5. Whenever you need to use the secret within your Golang program or
   script, call the **salainen** CLI/Get function passing in the JSON
   configuration file name and the custom config such as directory and
   private key, or the GPG settings.
6. The Get function returns the secret in the clear.

_Note_: It is important that you encrypt the JSON configuration file
before committing to Git. This can be part of your pipeline process.
You can use the **salainen** CLI Put function to achieve this.

## Command line

Set a secret with the default EJSON algorithm:

```text
salainen 'ejson:MY_SECRETS.json' 'secretvalue'
```

Set a secret with the GPG algorithm (after setting `"Algorithm": "GPG"`
in the provider configuration):

```text
salainen 'ejson:MY_SECRETS.json' 'secretvalue'
```

Retrieve a secret:

```text
salainen 'ejson:MY_SECRETS.json'
```

The recipient can be overridden for an individual file on write:

```text
salainen 'ejson:MY_SECRETS.json|other@example.com' 'secretvalue'
```

If the value after `|` does not match an existing GPG key, it is used as
the private key passphrase for that secret:

```text
salainen 'ejson:MY_SECRETS.json|my-passphrase' 'secretvalue'
```

## Golang Package

```go
salainen.Register("<config file>")
err := salainen.Set("ejson:<file>", "<value>")
secret_value, err := salainen.Get("ejson:<file>")
```
