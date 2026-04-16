# AGENTS

Guidance for AI coding assistants working in `dev-soprasteriano/gheese`. Read this file before making changes.

## Build, test, and release commands

- Build the CLI from the repository root with `go build -o gheese .`
- Run the full test suite with `go test ./...`
- Run a single test with `go test ./path/to/package -run TestName`
- Example for the current parser tests: `go test ./cmd -run TestParseRepositoryReference`
- There is no repo-specific lint command or linter config checked in
- Release artifacts are built in GitHub Actions from the repo root with `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ... .` and then packaged per OS/architecture

## High-level architecture

- `main.go` only calls `cmd.Execute()`
- `cmd/` contains the Cobra CLI surface:
  - `root.go` defines the root command
  - `repo.go` defines the `repo` command group
  - `list.go` and `move.go` register subcommands in `init()`
- `internal/github/` contains the GitHub API integration and command-facing business logic:
  - `client.go` builds the authenticated `go-github` client from `GITHUB_TOKEN`
  - `listRepos.go` wraps repository listing, normalizes the visibility filter, and paginates through all result pages before returning
  - `transferRepo.go` handles single-repo transfers, interactive bulk transfers, and the `y/n` confirmation loop used by `repo move --All`
- The command layer is intentionally thin: parse Cobra args/flags, create the client, call `internal/github`, and print results/errors

## Key conventions

- Authentication is environment-only: GitHub access always comes from `GITHUB_TOKEN`; there is no config file or alternate auth flow in the codebase
- Add new GitHub behavior under `internal/github` first, then wire it into `cmd/`; avoid putting API logic directly in Cobra command handlers
- Preserve the existing CLI contract unless the user explicitly wants it changed. In particular, `repo move` uses `--Source` and `--Destination`; single-repo moves expect `org/repo`, while `--All` expects plain organization names
- Keep user-input normalization close to the API wrapper. `normalizeVisibilityFilter` in `internal/github/listRepos.go` is the existing pattern for validating/normalizing CLI values before they reach `go-github`
- `ListRepos` is the shared repository inventory helper for both `repo list` and bulk move. If you change listing behavior, keep full pagination working for both paths
- For `repo move`, parsing and validation of `--Source` / `--Destination` lives in `cmd/move.go`; keep that mode-aware parsing centralized instead of duplicating split/validation logic
- `repo move --All` is interactive by design: it lists repos from the source org, prompts on stdin for each one, and only counts requested transfers, so changes to bulk move behavior must preserve that flow unless intentionally redesigning it
- README examples are the main user-facing command documentation. Update `README.md` when flags, command behavior, or installation/build steps change
- The repo uses DCO sign-off (`git commit -s`) for contributions
- Releases are managed with Release Please (`release-please-config.json` and `.release-please-manifest.json`), and conventional commit types feed the generated changelog sections

## Security

- Keep secrets secret. Do not expose any sensitive information in logs, outputs, or any other accessible forms.
- Do not edit files outside this repository unless the user explicitly asks for that.
- Prefer the Go standard library, existing internal packages, Cobra, and the current `go-github` integration. Avoid adding new third-party dependencies unless they are clearly necessary.
