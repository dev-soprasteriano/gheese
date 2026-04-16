# AGENTS

Guidance for AI coding assistants working in `dev-soprasteriano/gheese`. Read this file before making changes.

## How to work in this repo

- Keep changes surgical. This is a small CLI, so prefer focused edits over broad refactors.
- Put GitHub API behavior in `internal/github`, then wire it into `cmd/`. Keep Cobra command handlers thin.
- Preserve the current CLI contract unless the task explicitly changes it:
  - `repo list` uses `--visibility`
  - `repo move` uses `--Source` / `--Destination`
  - `repo move --All` accepts organization names and stays interactive
- Keep validation close to the helper that owns it. In particular, `cmd/move.go` owns parsing and validation of `--Source` / `--Destination`.
- Update `README.md` whenever command flags, command behavior, or installation/build steps change.
- Use `GITHUB_TOKEN` for authentication. Do not introduce alternate auth or config flows unless explicitly requested.
- Prefer the Go standard library, existing internal packages, Cobra, and `go-github`. Avoid adding new third-party dependencies unless they are clearly necessary.
- Do not edit files outside this repository unless the user explicitly asks for that.
- Do not expose secrets in logs, output, or committed files.

## Commit attribution

- This repo uses DCO sign-off (`git commit -s`) for human contributors, but AI assistants must not add `Signed-off-by` trailers themselves.
- AI assistants must not add `Co-authored-by` trailers for themselves.
- If an AI assistant creates or prepares a commit, disclose that assistance with an `Assisted by` trailer that names the agent and model.
- Example: `Assisted by: GitHub Copilot (GPT-5.4)`

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

- Keep user-input normalization close to the API wrapper. `normalizeVisibilityFilter` in `internal/github/listRepos.go` is the existing pattern for validating/normalizing CLI values before they reach `go-github`
- `ListRepos` is the shared repository inventory helper for both `repo list` and bulk move. If you change listing behavior, keep full pagination working for both paths
- `repo move --All` is interactive by design: it lists repos from the source org, prompts on stdin for each one, and only counts requested transfers, so changes to bulk move behavior must preserve that flow unless intentionally redesigning it
- Releases are managed with Release Please (`release-please-config.json` and `.release-please-manifest.json`), and conventional commit types feed the generated changelog sections
