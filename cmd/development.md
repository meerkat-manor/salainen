# Development of Salainen CLI

If you are developing and enhancing the **salainen** code the following
go execution commands might be useful.

## Providers

List your current providers

On Windows,

```bat
go run .\cmd\main.go -provider
```

On Linux,

```bash
go run ./cmd/main.go -provider
```

## Test keyring

Set and get keyring values

On Windows,

```bat
go run .\cmd\main.go keyring:tester01 my_secret
go run .\cmd\main.go keyring:tester01
```

On Linux,

```bash
go run ./cmd/main.go keyring:tester01 my_secret
go run ./cmd/main.go keyring:tester01
```

## Using a configuration file

Set and get keyring values 

On Windows,

```bat
go run .\cmd\main.go -config .\tests\configs\test_config01.json keyring:tester01 my_secretC
go run .\cmd\main.go -config .\tests\configs\test_config01.json keyring:tester01
# Or alternatively
$env:salainen=".\tests\configs\test_config01.json"
go run .\cmd\main.go keyring:tester01
```

On Linux,

```bash
go run ./cmd/main.go -config ./tests/configs/test_config01.json keyring:tester01 my_secretC
go run ./cmd/main.go -config ./tests/configs/test_config01.json keyring:tester01
# Or alternatively
export SALAINEN=".\tests\configs\test_config01.json"
go run ./cmd/main.go keyring:tester01
```

## Test secret generation

Set and get keyring values based on generated secret

On Windows,

```bat
go run .\cmd\main.go -generate keyring:tester02
go run .\cmd\main.go keyring:tester02
```

On Linux,

```bash
go run ./cmd/main.go -generate keyring:tester02
go run ./cmd/main.go keyring:tester02
```
