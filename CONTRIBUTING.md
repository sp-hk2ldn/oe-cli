# Contributing

Thanks for contributing to `oe-cli`.

## Development setup
1. Install Go (current project uses `go.mod` module mode).
2. Clone the repository.
3. Run:
```bash
go build ./...
go test ./...
```

## Local checks before opening a PR
Run these from repository root:
```bash
gofmt -w .
go vet ./...
go test ./...
```

Optional parity check (when Swift repo exists nearby):
```bash
./scripts/parity_check.sh
```

## Pull request guidelines
1. Keep changes focused and small where possible.
2. Update docs when command behavior changes:
   - `README.md`
   - `docs/COMMANDS.md`
3. Add or update tests/golden fixtures for CLI output changes.
4. Explain user impact clearly in PR description.

## Commit style
- Use clear imperative commit messages.
- Include scope when helpful, for example:
  - `apps: add eligibility command`
  - `docs: update command reference`

## Reporting bugs and requesting features
- Open a GitHub issue with reproduction steps and expected behavior.
- For security-sensitive findings, follow `SECURITY.md`.
