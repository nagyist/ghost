# Ghost CLI

A command-line interface for managing PostgreSQL databases.

## Installation

Multiple installation methods are provided. If you aren't sure, use the first one.

### Install Script (macOS/Linux/WSL)

```
curl -fsSL https://install.ghost.build | sh
```

### Install Script (Windows PowerShell)

```powershell
irm https://install.ghost.build/install.ps1 | iex
```

### Debian/Ubuntu

```bash
curl -s https://packagecloud.io/install/repositories/timescale/ghost/script.deb.sh | sudo os=any dist=any bash
sudo apt-get install ghost
```

### Red Hat/Fedora

```bash
curl -s https://packagecloud.io/install/repositories/timescale/ghost/script.rpm.sh | sudo os=rpm_any dist=rpm_any bash
sudo yum install ghost
```

### npm

```bash
npm install -g @ghost.build/cli
```

## Usage

```bash
ghost login    # Authenticate with GitHub OAuth
ghost create   # Create a new Postgres database
ghost list     # List all databases
```

## Commands

| Command | Description |
|---------|-------------|
| `api-key` | Manage API keys (create, list, delete) |
| `completion` | Generate the autocompletion script for the specified shell |
| `config` | Manage CLI configuration |
| `connect` | Get connection string for a database |
| `create` | Create a new Postgres database |
| `delete` | Delete a database |
| `feedback` | Send feedback to the Ghost team |
| `fork` | Fork a database |
| `help` | Help about any command |
| `list` | List all databases |
| `logs` | View logs for a database |
| `login` | Authenticate with GitHub OAuth |
| `logout` | Remove stored credentials |
| `mcp` | Ghost Model Context Protocol (MCP) server |
| `password` | Update the password for a database |
| `psql` | Connect to a database using psql |
| `rename` | Rename a database |
| `resume` | Resume a paused database |
| `schema` | Display database schema information |
| `sql` | Execute SQL query on a database |
| `status` | Show space usage |
| `version` | Show version information |

Run `ghost [command] --help` for more information about a command.

## Deployment

Releases are automatic on tag push. When you create a new GitHub Release with a semver tag (e.g. `v0.1.0`), the `Release` workflow runs GoReleaser to build and publish binaries, Docker images, and Linux packages.

To create a release, go to the repo's [Releases](../../releases) page and click **Draft a new release**, or use the GitHub CLI:

```bash
gh release create v<VERSION> --generate-notes --latest
```
