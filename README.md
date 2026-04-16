# gheese

`gheese` is a Go CLI from **Sopra Steria AS** for working more efficient with GitHub from the command line.

> Note: This tool is under active development. Until the release of v1 there can be major breaking changes and potential bugs in every release. Sopra Steria will not take any responsibility for any issues you could face by using this tool.
> Note 2: Feature requests and bug reports are very welcome, but be aware that it might take time to implement/fix these.

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
- Releases are built for macOS, Linux, and Windows through GitHub Actions.

## Contributing

This repository uses the **Developer Certificate of Origin (DCO)**. By contributing, you agree to the terms in [`DCO`](./DCO).

General contribution guidance is available in [`CONTRIBUTOR`](./CONTRIBUTOR).

Sign off commits with:

```bash
git commit -s
```

## Security

For severe vulnerabilities, follow the reporting instructions in [`SECURITY`](./SECURITY).
