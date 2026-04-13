# gheese

`gheese` is a Go CLI from **Sopra Steria AS** for working with GitHub repositories from the command line.

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

Replace `<version>` below with the release version you want, without the leading `v`.

### macOS and Linux

Download the matching archive, extract it, and move the binary somewhere on your `PATH`:

```bash
curl -LO https://github.com/<owner>/<repo>/releases/download/v<version>/gheese_<version>_linux_amd64.tar.gz
tar -xzf gheese_<version>_linux_amd64.tar.gz
chmod +x gheese
sudo mv gheese /usr/local/bin/gheese
```

For macOS, use the matching `darwin` archive instead of `linux`. For Apple Silicon, use the `arm64` artifact.

### Windows

Download the matching `.zip` asset from the release page, extract `gheese.exe`, and place it in a directory on your `PATH`.

In PowerShell:

```powershell
Invoke-WebRequest -OutFile gheese.zip https://github.com/<owner>/<repo>/releases/download/v<version>/gheese_<version>_windows_amd64.zip
Expand-Archive gheese.zip -DestinationPath .
```

After extraction, move `gheese.exe` to a folder that is included in `PATH`.

## Commands

### `repo list`

Lists the repositories in an organization that the authenticated user can see.

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
  -o source-org \
  -r source-repo \
  -O destination-org \
  -R destination-repo
```

Useful flags:

| Flag | Meaning |
| --- | --- |
| `-o`, `--SourceOrganization` | Source organization |
| `-r`, `--SourceRepository` | Source repository |
| `-O`, `--DestinationOrganization` | Destination organization |
| `-R`, `--DestinationRepository` | Repository name after transfer |
| `-t`, `--TeamId` | Team IDs to add to the repository |

### `repo move --All`

Walks through all repositories in a source organization and asks for confirmation before each transfer.

```bash
./gheese repo move \
  -o source-org \
  -O destination-org \
  -A
```

What it does:

- lists repositories in the source organization
- prompts `y/n` for each repository
- requests a transfer for each repository you approve
- prints how many repositories were processed and how many transfer requests were made

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
