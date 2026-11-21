# Encrypted File secret storage

The **efile** provider stores each secret in a separate file. The
contents are **encrypted** using the configured key and encryption
algorithm.

If no algorithm is specified it defaults to `ChaCha20-Poly1305`.

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

## Supported encryption algorithms

The following algorithms are supported:

* `ChaCha20-Poly1305` (default)
* `GPG` / `OpenPGP`

When `GPG` is selected the secret is encrypted with the public key of
the configured recipient. The matching private key is required to read
the secret.

## Configuration

| Setting | Description |
| --- | --- |
| `RootPath` | Directory where secret files are stored. |
| `Algorithm` | `ChaCha20-Poly1305` (default) or `GPG`. |
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

A GPG configuration example:

```json
        "efile": {
            "enabled": true,
            "name": "Encrypted File System (GPG)",
            "custom": {
                "RootPath": "~/.secrets/salainen/efile",
                "Algorithm": "GPG",
                "GPGHomeDir": "~/.secrets/salainen/gnupg",
                "GPGRecipient": "me@example.com"
            }
        },
```

_Note_: `GPGRecipient` can be omitted after `-init` has generated a key.
The recipient is then read from the marker file in `RootPath`
(`.salainen-efile-gpg.json`).

## Setup

### Install GPG

#### Linux

Install GnuPG with your distribution package manager. For Debian/Ubuntu:

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

Install Gpg4win, which includes GPG and the Kleopatra key manager. You
can download it from [https://www.gpg4win.org/](https://www.gpg4win.org/)
or install it with `winget`:

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
the provider is already initialised before doing anything. If it is not
initialised, you are prompted for the key name, email address,
passphrase and expiry.

```text
salainen -provider -init efile
```

The `-init` option may also be placed after the provider name:

```text
salainen -provider efile -init
```

The generated key is written to the GPG keyring (`GPGHomeDir` when set)
and a marker file is written to `RootPath`. Running `-init` again reports
that the provider is already initialised.

For non-interactive setup, supply the values in the provider `custom`
configuration:

```json
            "custom": {
                "RootPath": "~/.secrets/salainen/efile",
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

You can also use an existing key or create one with GPG directly.

#### GPGP setup on Linux

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

## Command line

Set a secret:

```text
salainen 'efile:MY_SECRET' 'secretvalue'
```

Retrieve a secret:

```text
salainen 'efile:MY_SECRET'
```

The value is written to standard output. If there is an error the
program exits with a non-zero status.

The recipient can be overridden for an individual file on write:

```text
salainen 'efile:MY_SECRET|other@example.com' 'secretvalue'
```

The file is encrypted for `other@example.com`; decryption still uses
the private key available in the local GPG keyring.

If the value after `|` does not match an existing GPG key, it is used as
the private key passphrase for that secret:

```text
salainen 'efile:MY_SECRET|my-passphrase' 'secretvalue'
```

## Golang Package

To set the secret value you call the function in your Go code as:

```go
salainen.Set("efile:<key>", "<value>")
```

This will use the default configuration and without
a `salainen.json` or `salainen.yml` being in the current file
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
secret_value, err := salainen.Get("efile:<key>")
```
