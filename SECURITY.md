# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| `v0.0.10` and newer | Yes |
| Older releases | No |

This project is pre-1.0. Security fixes are made on the latest tagged release and `main`; users should upgrade to the newest available version.

## Reporting a vulnerability

Please report suspected vulnerabilities privately.

- Use [GitHub private vulnerability reporting](https://github.com/meerkat-manor/salainen/security/advisories/new).
- Do not open a public issue, discussion, or pull request for an unfixed vulnerability.
- If private reporting is unavailable, contact the maintainers through the project's GitHub repository without disclosing exploit details publicly.

A useful report should include:

- affected version, commit, or release
- affected provider, component, or configuration
- impact and severity assessment
- reproduction steps or a minimal proof of concept
- suggested fix, if known

Response targets:

- acknowledgement within 5 business days
- status update at least every 10 business days
- coordinated disclosure after a fix and release are available

Please allow time for investigation and remediation before public disclosure.

## Security model

Salainen provides a consistent interface over secret providers; it does not replace a full secrets platform or operating-system isolation.

Configuration files are trusted inputs. Anyone able to write the active configuration can redirect secret lookups, configure executable paths, or change provider behavior. Protect configuration files with restrictive permissions and limit write access.

Provider-specific trust assumptions include:

- **file, efile, ejson, keepass:** local filesystem access to the secret files or derived keys can expose secrets.
- **gpg:** the configured GPG executable must be an absolute path, but configuration remains trusted. GPG passphrases are supplied through temporary `0600` files, not command-line arguments.
- **env, plain, promptsec:** these providers intentionally expose secret material through environment variables, literal configuration, or terminal output.
- **vault, bitwarden:** remote endpoints require HTTPS except loopback addresses. Credentials embedded in endpoint URLs are rejected.
- **debug mode:** debug output may include resolved configuration paths; do not enable it in shared or public environments. It is only intended for development.

## Known residual risks

The following items are documented as residual risks and are intentionally as part of the design:

- stdout, clipboard, and environment exposure
- indirect cross-provider reference cycles
- cross-platform ownership enforcement for discovered configuration files

## Safe usage guidance

- Keep configuration files and secret stores readable only by the account running Salainen.
- Prefer external secret managers over local encrypted files for high-value secrets.
- Disable providers that are not needed.
- Do not store production secrets in `plain`, `env`, or unencrypted `file` providers.
- Do not commit secret files, configuration files containing credentials, or generated key material to source control.
- Run the latest release and subscribe to repository security advisories.

## Scope

In scope:

- the `salainen` Go package and CLI
- provider implementations under `extensions/`
- configuration loading and provider registration
- secret generation, file permissions, path validation, and endpoint validation

Out of scope:

- compromise of the host operating system, container runtime, or user account
- attacks requiring write access to trusted configuration files
- risks inherent to explicitly plaintext providers
- third-party services and their clients except for Salainen's integration code
- social engineering or physical access
