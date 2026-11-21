# Salainen Developer Guide

This guide describes how to build, test and extend **salainen** locally.

- Module: `github.com/meerkat-manor/salainen`
- Go version: `1.23.0` (toolchain `1.23.4` in `go.mod`); the suite is also verified with Go 1.26.8.
- Tests: standard `testing` package, no external Go test framework.
- Optional: Docker, used to launch a real HashiCorp Vault for integration tests (see [Vault Integration Testing](VAULT-INTEGRATION-TESTING.md)).

---

## 1. Repository layout

| Path | Purpose |
|---|---|
| `salainen.go` | Public API: `SecretStorage` interface, provider registry, `Get`/`Put` and key parsing. |
| `config/` | Configuration discovery, JSON/YAML parsing and provider registration. |
| `cmd/` | CLI entry point (`main.go`) and CLI-specific recipes in `development.md`. |
| `extensions/<name>/` | One package per provider (`env`, `file`, `encryptedfile`, `ejsons`, `keyring`, `wincred`, `keepass`, `bitwarden`, `vault`, `plain`, `promptsec`) plus the shared `gpgutil` helper. |
| `extensions/bitwarden/genbw/` | Generated OpenAPI client; do not edit by hand. |
| `generate/` | Credential/secret generation used by the CLI `-generate` flag. |
| `sync/` | Placeholder for a future feature. |
| `tests/` | Static fixtures (configuration files, KeePass databases, EJSON samples). |
| `docs/` | This guide, the Vault integration guide and the security/enhancement report. |
| `.well-known/.woodpecker.yml` | CI definition (see the known issue in section 8). |

---

## 2. Common commands

```bash
# Build the CLI
go build -o salainen ./cmd/main.go

# Run every test (Vault tests start a Docker container when Docker is available)
go test ./...

# Verbose output for one package
go test ./extensions/vault/... -v -count=1

# Race detector
go test -race ./...

# Static checks
go vet ./...
gofmt -l .

# Coverage overview
go test -cover ./...
```

`-count=1` disables the test result cache and is useful when diagnosing integration failures.

---

## 3. Test structure and conventions

Tests live next to the code they exercise. Existing suites include:

- `salainen_test.go`, `salainen_more_test.go` — core API and key parsing.
- `config/*_test.go` — config loading, discovery, debug behaviour.
- `extensions/*/..._test.go` — provider-specific unit tests.
- `extensions/vault/integration_test.go` — Docker-backed Vault harness and integration cases.

Conventions to follow when adding tests:

1. Prefer `t.TempDir()` over writing into the user's home directory (`~/.secrets`). This keeps tests isolated on all platforms.
2. Use `t.Cleanup` to remove entries from the global `salainen.SecretStorages` registry after registering a provider.
3. Use `errors.Is` for sentinel errors (`salainen.ErrNoSuchSecret`, `ErrInvalidSecret`, `ErrInvalidSecretAccess`, `ErrInvalidProvider`).
4. Use table-driven subtests for validation matrices.
5. Never commit real credentials. The Vault harness uses the dev root token `salainen-root`; test KeePass fixtures use the trivial passwords documented in their README.
6. Tests that require external infrastructure must skip cleanly when it is unavailable rather than failing the build.

---

## 4. Vault integration tests (Docker)

The Vault package automatically starts a real Vault dev server in a Docker container via `TestMain` in `extensions/vault/integration_test.go`. No manual setup is required when Docker is installed.

Full details, environment variables and troubleshooting are in [Vault Integration Testing](VAULT-INTEGRATION-TESTING.md). Short version:

```bash
go test ./extensions/vault/... -v -count=1
SALAINEN_SKIP_DOCKER=1 go test ./extensions/vault/...   # skip server-dependent cases
SALAINEN_VAULT_ADDR=http://127.0.0.1:8200 go test ./extensions/vault/...
```

---

## 5. Adding a new secret provider

1. Create `extensions/<name>/<name>.go` with an unexported struct implementing `salainen.SecretStorage`:

   ```go
   type f struct{ /* configuration */ }

   func (sl *f) Help()                            {}
   func (sl *f) Init(custom interface{}) error    { return nil }
   func (sl *f) Get(path string) (string, error)  { /* ... */ }
   func (sl *f) Put(path, value string) error     { /* ... */ }
   ```

2. Export constructors `New(config string, custom interface{}) (salainen.SecretStorage, error)` and `Register(config string, custom interface{}) error`. `Register` calls `salainen.AddSecretStorage`.

3. Add the provider to the default provider list and/or sample configurations (`defaults.json`, `defaults.yaml`, `all_providers.*`) in `config/configuration.go`.

4. Optionally implement `salainen.SecretStorageInitializer` (`Initialised`/`Initialise`) for providers that need one-time setup, which the CLI `-init` flag can call.

5. Follow the path convention: split the key on `|` and treat the suffix as a provider-specific parameter (passphrase, recipient, etc.).

6. Add unit tests in the same package plus documentation in `extensions/<name>/README.md`.

7. Validate all `custom` values with the comma-ok type assertion form; never assume `map[string]string`.

---

## 6. Regenerating the Bitwarden client

The Bitwarden provider uses code generated from `bitwarden-management-api.json`.

- Script: `extensions/bitwarden/generate_client.ps1`
- Generator: `oapi-codegen` v2.3.0
- Configs: `types.cfg.yaml`, `client.cfg.yaml`
- Output: `genbw/general-type.gen.go`, `genbw/general-client.gen.go`

The config files declare `package: gen`, while the generated files are checked in as `package genbw`; preserve that rename after regeneration.

---

## 7. CLI development recipes

See `cmd/DEVELOPMENT.md` for ready-to-run examples covering provider listing, keyring round trips, configuration files and `-generate`. On Windows use `go run .\cmd\main.go`; on Linux/macOS use `go run ./cmd/main.go`.

---

## 8. Known issues and CI notes

- The CI pipeline in `.well-known/.woodpecker.yml` runs `cd src` in every step, but this repository has no `src/` directory. The pipeline needs correcting before it can pass.
- CI pins `golangci-lint` v1.50.1, which predates Go 1.23; upgrade it when enabling the lint step locally.
- `gofmt -l .` reports pre-existing formatting differences in several checked-in files. New and modified files should be gofmt-clean.
- Security and hardening findings are tracked in [security-and-enhancements-report.md](security-and-enhancements-report.md).

---

## 9. Troubleshooting

| Symptom | Likely cause / fix |
|---|---|
| Vault tests report "Docker executable not found" | Install Docker or set `SALAINEN_SKIP_DOCKER=1`. |
| Vault tests report "unable to start Vault container" | Check `docker info`; ensure the daemon is running and the image can be pulled. Override the image with `SALAINEN_VAULT_IMAGE` if needed. |
| Tests pass but Vault cases are `SKIP` | Docker was unavailable or explicitly disabled; use `-v` to see the reason. |
| `setenv: The parameter is incorrect` in config tests | Caused by setting an environment variable with an empty name; use the `ENV_SALAINEN_*` constants. |
| Port 8200 is already in use | The harness uses a random host port, so this should not matter. If you set `SALAINEN_VAULT_ADDR`, ensure that server is reachable and healthy. |
