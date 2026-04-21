---
title: gheese
---

# gheese

`gheese` is a Go CLI from **Sopra Steria AS** for working more efficient with GitHub repositories and user access from the command line.

> Note: This tool is under active development. Until the release of v1 there can be major breaking changes and potential bugs in every release. Sopra Steria will not take any responsibility for any issues you could face by using this tool.
> Note 2: Feature requests and bug reports are very welcome, but be aware that it might take time to implement/fix these.

## Additional documentation

- [Contributing](../CONTRIBUTOR)
- [Security](../SECURITY)

## Requirements

- Go installed locally if you want to build from source
- A GitHub personal access token available as `GITHUB_TOKEN`

```bash
export GITHUB_TOKEN=your-token
```

On Windows PowerShell:

```powershell
$env:GITHUB_TOKEN="your-token"
```

## Build

```bash
go build -o gheese .
```

## Install from a release

Each GitHub release includes prebuilt artifacts named like:

- `gheese_<version>_darwin_amd64.tar.gz`
- `gheese_<version>_darwin_arm64.tar.gz`
- `gheese_<version>_linux_amd64.tar.gz`
- `gheese_<version>_linux_arm64.tar.gz`
- `gheese_<version>_windows_amd64.zip`
- `gheese_<version>_windows_arm64.zip`

### macOS and Linux

Resolve the latest release tag, download the matching archive, extract it, and move the binary somewhere on your `PATH`:

```bash
VERSION="$(curl -fsSL https://api.github.com/repos/dev-soprasteriano/gheese/releases/latest | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')"
curl -fLO "https://github.com/dev-soprasteriano/gheese/releases/download/v${VERSION}/gheese_${VERSION}_linux_amd64.tar.gz"
tar -xzf "gheese_${VERSION}_linux_amd64.tar.gz"
chmod +x gheese
sudo mv gheese /usr/local/bin/gheese
```

For macOS, use the matching `darwin` archive instead of `linux`. For Apple Silicon, use the `arm64` artifact.

### Windows

Download the matching `.zip` asset from the release page, extract `gheese.exe`, and place it in a directory on your `PATH`.

In PowerShell:

```powershell
$version = (Invoke-RestMethod https://api.github.com/repos/dev-soprasteriano/gheese/releases/latest).tag_name.TrimStart("v")
Invoke-WebRequest -OutFile gheese.zip "https://github.com/dev-soprasteriano/gheese/releases/download/v$version/gheese_${version}_windows_amd64.zip"
Expand-Archive gheese.zip -DestinationPath .
```

After extraction, move `gheese.exe` to a folder that is included in `PATH`.

## Commands

### `user add`

Invites a user to one or more organizations and queues the requested team assignments on those invitations.

If you only know the user's email address, `user add` can create the organization invitations without `--login`.

Simple flag-driven example:

```bash
./gheese user add \
  --enterprise my-enterprise \
  --email octocat@example.com \
  --org platform \
  --team platform/core
```

Useful flags:

| Flag | Meaning |
| --- | --- |
| `--enterprise` | Enterprise slug that owns the cost center |
| `--login` | GitHub login for the user to invite |
| `--email` | Email address for the user. `user add` can use this without `--login` |
| `--org` | Target organization, repeatable |
| `--team` | Team in `org/team-slug` form, repeatable |
| `--role` | Invitation role: `direct_member`, `admin`, or `billing_manager` |
| `--file` | Path to a JSON request file |
| `--dry-run` | Validate and report actions without changing GitHub |
| `--output` | Output format: `text` or `json` |

Use a JSON file for automation:

```bash
./gheese user add --file onboarding-request.json --output json
```

Example request file:

```json
{
  "enterprise": "my-enterprise",
  "email": "octocat@example.com",
  "organizations": [
    {
      "name": "platform",
      "teams": ["core", "developers"]
    },
    {
      "name": "data",
      "teams": ["analysts"]
    }
  ]
}
```

### `user update`

Reconciles an existing user back to the requested state for organization membership, team membership, and cost center assignment. This command is designed for recurring workflows.

`user update` requires the user's GitHub login and cost center so it can verify membership, ensure team membership, and assign the cost center resource after invitation acceptance.

```bash
./gheese user update \
  --enterprise my-enterprise \
  --login octocat \
  --cost-center Engineering \
  --org platform \
  --team platform/core
```

### `repo list`

Lists the repositories in an organization that the authenticated user can see.

The command paginates through GitHub results, so it returns all visible repositories in the organization, not just the first page.

```bash
./gheese repo list my-organization
```

Filter by visibility:

```bash
./gheese repo list my-organization --visibility public
./gheese repo list my-organization --visibility private
```

`--visibility` accepts `all`, `public`, or `private`. If you do not set it, the command lists all visible repositories.

### `repo move`

Transfers a repository to another organization.

Example:

```bash
./gheese repo move \
  -s source-org/source-repo \
  -d destination-org/destination-repo
```

Useful flags:

| Flag | Meaning |
| --- | --- |
| `-s`, `--Source` | Source repository in `org/repo` form |
| `-d`, `--Destination` | Destination repository in `org/repo` form |
| `-t`, `--TeamId` | Team IDs to add to the repository |

### `repo move --All`

Walks through all repositories in a source organization and asks for confirmation before each transfer.

```bash
./gheese repo move \
  -s source-org \
  -d destination-org \
  -A
```

What it does:

- lists repositories in the source organization
- prompts `y/n` for each repository
- requests a transfer for each repository you approve
- prints how many repositories were processed and how many transfer requests were made

Like `repo list`, the bulk move flow paginates through the full repository list for the source organization before prompting.

## Notes

- The tool only works against repositories visible to the authenticated user.
- Repository transfers depend on the permissions and rules enforced by GitHub.
- `user add` requires organization-level invitation permissions. `user update` additionally requires enterprise billing permissions for cost center assignment. The default GitHub Actions token is not sufficient for these commands.
- Releases are built for macOS, Linux, and Windows through GitHub Actions.

## Contributing

This repository uses the **Developer Certificate of Origin (DCO)**. By contributing, you agree to the terms in the [DCO](https://github.com/dev-soprasteriano/gheese/blob/main/DCO).

General contribution guidance is available in [Contributing](../CONTRIBUTOR).

Sign off commits with:

```bash
git commit -s
```

## Security

For severe vulnerabilities, follow the reporting instructions in [Security](../SECURITY).
