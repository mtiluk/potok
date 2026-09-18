# Contributing to Potok

Thanks for helping improve Potok. Please keep pull requests focused on one
change and explain how you verified it.

## Development setup

Potok requires Go 1.24 or newer. Clone the repository and run the same checks
used by CI:

```bash
go build ./...
go test ./...
```

The client is built from `cmd/potok`; the server is built from `cmd/potokd`.

## Making a change

1. Fork the repository and create a topic branch from `main`.
2. Make the smallest change that solves the issue.
3. Add or update tests when behaviour changes.
4. Run `go build ./...` and `go test ./...` locally.
5. Open a pull request describing the change and the checks you ran.

Please do not commit credentials, local configuration files, generated
artifacts, or vault contents.
