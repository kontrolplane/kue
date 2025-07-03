# Contributing to Kue

First off, thank you for taking the time to contribute! 🙌

The following is a set of guidelines for contributing to Kue, which is hosted in the [kontrolplane](https://github.com/kontrolplane) organization on GitHub.

## Development Environment Setup

1. Clone the repo and set it as safe:

   ```bash
   git clone https://github.com/kontrolplane/kue.git
   cd kue
   go run ./...
   ```

2. Make sure `go vet`, `go test ./...`, and `golangci-lint run` pass.

## Pull Request Process

1. Fork / create a branch (`feature/xyz`).
2. Keep commits small and focused.
3. Add tests for new functionality where possible.
4. Ensure `make test` (or equivalent) runs without errors.
5. Reference an issue in the PR description (`Fixes #X`).

## Code of Conduct

This project uses the [Contributor Covenant](https://www.contributor-covenant.org/) code of conduct. By participating, you are expected to uphold this code.
