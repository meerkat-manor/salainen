# Vault Integration Testing with Docker

The Vault provider has two kinds of tests:

1. **Unit-style tests** that do not need a server (`TestVaultInitValidation`, `TestVaultRegister`, `TestSimpleVaultCreate`). They always run.
2. **Integration tests** that read and write real secrets through the HashiCorp Vault API. They use a Docker container started automatically by `TestMain` in `extensions/vault/integration_test.go`.

---

## 1. How the harness works

`TestMain(m *testing.M)` runs once for the whole `vault` package:

1. Resolves the target server:
   - If `SALAINEN_VAULT_ADDR` is set, uses that external server (no container is started).
   - Otherwise, if `SALAINEN_SKIP_DOCKER` is set, records an "unavailable" reason so integration tests skip.
   - Otherwise checks for `docker` on `PATH` and starts a container.

2. Starts the container with:

   ```text
   docker run --detach --rm \
     --cap-add IPC_LOCK \
     -e VAULT_DEV_ROOT_TOKEN_ID=salainen-root \
     -p 127.0.0.1::8200 \
     hashicorp/vault:1.17 server -dev
   ```

   The empty host side of `-p` asks Docker to allocate a free loopback port, so port 8200 on the host does not need to be free.

3. Discovers the published port with `docker port <id> 8200/tcp` (polled for up to 3 minutes to allow for a first-time image pull).

4. Polls `GET /v1/sys/health` until Vault reports an active/unsealed state (up to 90 seconds).

5. Enables a KV version 2 secrets engine at `salainen/` through `POST /v1/sys/mounts/salainen`. The dev server already provides KV v2 at `secret/`; the extra mount is used by `TestKV01VaultPutGet`.

6. Runs the tests.

7. Removes the container with `docker rm --force <id>` after `m.Run()` returns.

If any setup step fails, the error is stored and every server-dependent test calls `t.Skipf` with that reason. A failure to start Docker therefore never breaks `go test ./...` on a machine without Docker.

---

## 2. Running the tests

```bash
# Full suite; Vault integration tests use Docker automatically
go test ./...

# Just the Vault package, verbose, no result cache
go test ./extensions/vault/... -v -count=1

# Force a fresh image
docker pull hashicorp/vault:1.17
go test ./extensions/vault/... -count=1
```

PowerShell equivalents work identically; set environment variables with `$env:NAME = "value"`.

---

## 3. Configuration environment variables

| Variable | Default | Purpose |
|---|---|---|
| `SALAINEN_VAULT_ADDR` | *(unset)* | Use an existing Vault server instead of starting Docker, e.g. `http://127.0.0.1:8200`. The `salainen/` KV v2 mount is created on it if absent. |
| `SALAINEN_VAULT_IMAGE` | `hashicorp/vault:1.17` | Docker image to run. |
| `SALAINEN_VAULT_TOKEN` | `salainen-root` | Dev root token, used for the API calls and as the provider's `AccessToken`. |
| `SALAINEN_SKIP_DOCKER` | *(unset)* | Any non-empty value disables Docker auto-start; server-dependent tests skip. |

Example — using an externally managed server:

```bash
export SALAINEN_VAULT_ADDR=http://127.0.0.1:8200
export SALAINEN_VAULT_TOKEN=my-dev-token
go test ./extensions/vault/... -v
```

---

## 4. Integration test cases

| Test | What it verifies |
|---|---|
| `TestSimpleVaultPutGet` | Round trip in the default `secret/` KV v2 mount, plus an error for a missing key. |
| `TestKV01VaultPutGet` | Round trip in the custom `salainen/` KV v2 mount created by the harness. |
| `TestVaultPutOverwrite` | A second `Put` replaces the stored value. |
| `TestVaultPipeSuffixIsIgnored` | `key|parameter` uses `key` as the Vault path. |
| `TestVaultNestedPath` | Slash-delimited paths are stored as KV v2 sub-paths. |
| `TestVaultMissingSecret` | Reading a non-existent key returns an error. |
| `TestVaultCustomElementName` | `ElementName` selects the stored field; a storage with the default element cannot read it. |
| `TestVaultDefaultMount` | Omitting `Vault` falls back to `secret`. |
| `TestVaultInvalidToken` | An invalid token causes authentication errors on both `Put` and `Get`. |

Support tests that need no server: `TestVaultInitValidation`, `TestVaultRegister`, `TestSimpleVaultCreate`.

---

## 5. Writing new Vault integration tests

Use the shared helpers rather than hard-coding an address:

```go
func TestMyVaultCase(t *testing.T) {
    addr := requireVault(t)          // skips when the server is unavailable
    storage := newTestStorage(t, addr, map[string]string{
        "Vault": "salainen",
    })

    if err := storage.Put("my/key", "value"); err != nil {
        t.Fatalf("unexpected put error: %v", err)
    }
    value, err := storage.Get("my/key")
    if err != nil || value != "value" {
        t.Fatalf("unexpected result: %q %v", value, err)
    }
}
```

- `requireVault(t)` returns the base URL and skips on setup failure.
- `newTestStorage(t, addr, extra)` registers the `plain` provider, builds the custom configuration and constructs a storage. Pass overrides in `extra`.
- Call `Put` before a "missing key" `Get`: the provider sets its client token during `Put`, so this makes the assertion meaningful.
- Use `testInit()` directly only in tests that construct storages themselves.

---

## 6. Manual container (optional)

The harness is self-contained, but a long-lived server is useful for CLI experiments. The repository also ships a compose file:

```bash
docker compose -f extensions/vault/docker/docker-compose.yml up -d
export SALAINEN_VAULT_ADDR=http://127.0.0.1:8200
export SALAINEN_VAULT_TOKEN=salainen-root
go test ./extensions/vault/... -v
docker compose -f extensions/vault/docker/docker-compose.yml down
```

The compose file publishes the fixed host port 8200 and uses the same dev token. When `SALAINEN_VAULT_ADDR` is set, the harness will not start or stop a container, so the compose-managed server remains under your control.

---

## 7. Troubleshooting

| Symptom | Cause / action |
|---|---|
| `Vault integration server unavailable: docker executable not found` | Install Docker or set `SALAINEN_SKIP_DOCKER=1`. |
| `unable to start Vault container` | Check `docker info`, daemon status and network access for the image pull. Try `docker pull hashicorp/vault:1.17` first. |
| Timed out waiting for port mapping | Container exited early; inspect with `docker ps -a` and `docker logs <id>`. |
| Timed out waiting for Vault | Vault is sealed or unhealthy; check `docker logs <id>`. |
| 403 errors on `Get` | The client token is set by `Put`; write a value first, or supply a valid `SALAINEN_VAULT_TOKEN`. |
| Container left behind | Normally removed by `TestMain`. Clean up with `docker ps -a --filter ancestor=hashicorp/vault:1.17` and `docker rm -f <id>`. |
