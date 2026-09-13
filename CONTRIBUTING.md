# Contributing

Potok is in active development. PRs are welcome — keep them small and tested.

## Build and test

You need Go 1.25+ (see `go.mod`).

From the repo root:

```bash
go build ./...
go test ./...
```

Those are the same commands CI runs (`.github/workflows/ci.yml`).

To produce the two binaries:

```bash
go build -o potok ./cmd/potok     # CLI
go build -o potokd ./cmd/potokd   # server
```

`potokd` reads `ADDR`, `DATA_DIR`, and `DATABASE_URL` from the environment (or a `.env` file in the working directory). There is no Makefile.

Layout:

| Path | What |
|---|---|
| `cmd/potok` | CLI |
| `cmd/potokd` | Server |
| `internal/` | Client, server, and protocol code |
| `migrations/` | SQL migrations (embedded) |

## Code style

- `gofmt` everything before you push (`gofmt -w .`).
- Match the file you're in. This is ordinary Go: cobra commands under `internal/client/cli`, packages under `internal/`, wrapped errors (`fmt.Errorf("...: %w", err)`).
- Tests live next to the code (`*_test.go`) and use the standard `testing` package — no extra assertion libraries. Table-driven tests and `t.Helper()` are the usual pattern.
- Don't add a linter, Makefile, or test framework unless the change needs it.

## Pull requests

1. Fork the repo.
2. Create a branch off `main`.
3. Make the change. Add or update tests if behaviour changes.
4. Run `gofmt -w .` and `go test ./...`.
5. Open a PR against `main`.

Keep the PR focused. Say what changed and why. If it fixes an issue, mention it (`Fixes #58`).
