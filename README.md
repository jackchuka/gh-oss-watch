<div align="center">

# gh-oss-watch

**One terminal view of every open-source repo you maintain.**

A GitHub CLI extension that tracks stars, issues, PRs, forks, releases, and Dependabot alerts<br>
across your repositories, and shows only what changed since you last looked.

[![Test](https://github.com/jackchuka/gh-oss-watch/actions/workflows/test.yml/badge.svg)](https://github.com/jackchuka/gh-oss-watch/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/jackchuka/gh-oss-watch?sort=semver)](https://github.com/jackchuka/gh-oss-watch/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

<img src="assets/demo.gif" alt="gh oss-watch adding a repo, then showing the dashboard and release status across five repos" width="800">

[Installation](#installation) · [Quick Start](#quick-start) · [Commands](#commands) · [Security Alerts](#security-alerts) · [Configuration](#configuration)

</div>

---

Maintaining more than a handful of repositories means clicking through each one to find new issues, open PRs, merged work waiting for a release, and security alerts. **gh-oss-watch** keeps a watch list of your repos and answers those questions in one command each: what's new since the last check, which repos have unreleased commits, and where Dependabot has open alerts.

## Features

- **Multi-repo watch list** — track any number of repositories from a single YAML config
- **Diff-based status** — `status` reports only new stars, issues, PRs, forks, and releases since the previous run
- **Dashboard** — current counts, latest release, and last update for every watched repo, with totals
- **Release tracking** — see which repos have commits since their latest release
- **Security alerts** — severity-ranked summary of open Dependabot alerts across the watch list
- **Fans** — find the people who starred the most of your repos
- **Per-repo events** — choose which events matter for each repository
- **Scriptable** — `--format json` on every report command; concurrent API requests with configurable limits

## Installation

Requires the [GitHub CLI](https://github.com/cli/cli), installed and authenticated (`gh auth login`).

### GitHub CLI Extension

```bash
gh extension install jackchuka/gh-oss-watch
```

### From Source

```bash
git clone https://github.com/jackchuka/gh-oss-watch.git
cd gh-oss-watch
go build -o gh-oss-watch .
```

## Quick Start

1. **Create the config file:**

   ```bash
   gh oss-watch init
   ```

2. **Add repositories** (all events except `releases` are tracked by default):

   ```bash
   gh oss-watch add owner/repo
   gh oss-watch add owner/other-repo stars issues releases
   ```

3. **Check what's new** since the last run:

   ```bash
   gh oss-watch status
   ```

4. **See the full picture:**

   ```bash
   gh oss-watch dashboard
   gh oss-watch releases
   ```

## Commands

| Command                  | Description                                     | Example                                    |
| ------------------------ | ----------------------------------------------- | ------------------------------------------ |
| `init`                   | Create the config file                          | `gh oss-watch init`                        |
| `add <repo> [events...]` | Add a repo to the watch list                    | `gh oss-watch add owner/repo stars issues` |
| `set <repo> <events...>` | Change the events tracked for a repo            | `gh oss-watch set owner/repo forks`        |
| `remove <repo>`          | Remove a repo from the watch list               | `gh oss-watch remove owner/repo`           |
| `list`                   | List watched repos (`-l, --lang` to filter)     | `gh oss-watch list --lang go`              |
| `status`                 | Show new activity since the last check          | `gh oss-watch status`                      |
| `dashboard`              | Show current counts across all repos            | `gh oss-watch dashboard`                   |
| `releases`               | Show release status (`-u, --only-unreleased`)   | `gh oss-watch releases -u`                 |
| `security`               | Show open Dependabot alerts across repos        | `gh oss-watch security --detail`           |
| `fans`                   | Show who starred your repos (`-t, --top N`)     | `gh oss-watch fans --top 10`               |

Global flags:

| Flag               | Description                    | Default |
| ------------------ | ------------------------------ | ------- |
| `-f, --format`     | Output format: `plain`, `json` | `plain` |
| `--max-concurrent` | Max concurrent API requests    | `10`    |
| `--timeout`        | Request timeout in seconds     | `30`    |

### Event Types

| Event           | Tracks                                              |
| --------------- | --------------------------------------------------- |
| `stars`         | Increase in stars                                   |
| `issues`        | Increase in open issues                             |
| `pull_requests` | Increase in open pull requests                      |
| `forks`         | Increase in forks                                   |
| `releases`      | New releases, and commits since the latest release  |

## Security Alerts

`gh oss-watch security` scans the watch list for open Dependabot alerts and prints a severity-ranked snapshot. Unlike `status`, it is not a diff: an open alert stays listed until it is fixed. Repos whose alerts you can't read (not owned by you, or alerts disabled) are listed as skipped.

```bash
gh oss-watch security                       # severity-ranked summary table
gh oss-watch security --detail              # every alert (package, range, fix, GHSA)
gh oss-watch security --severity high       # only high and critical
gh oss-watch security --repo owner/repo     # a single watched repo
gh oss-watch security --format json         # machine-readable
```

## Configuration

The watch list lives in `~/.gh-oss-watch/config.yaml`. The last-seen counts used by `status` are cached next to it in `cache.yaml`.

```yaml
repos:
  - repo: owner/repo
    events:
      - stars
      - issues
      - pull_requests
      - releases
    language: Go
  - repo: owner/other-repo
    events:
      - stars
      - forks
```

`language` is filled in from GitHub when you `add` a repo and is used by `list --lang`.

## Automation

`status` keeps its own state, so it works well on a schedule:

```bash
# Check for new activity every hour
0 * * * * gh oss-watch status

# Weekly summary by mail
0 9 * * 1 gh oss-watch dashboard | mail -s "Weekly OSS summary" you@example.com
```

## Development

```bash
go build -o gh-oss-watch .
go test -race ./...
golangci-lint run
cd services && go generate   # regenerate mocks
```

The demo GIF is recorded with [VHS](https://github.com/charmbracelet/vhs): `vhs assets/demo.tape`.

## Contributing

Contributions are welcome. For larger changes, please open an issue first to discuss what you would like to change, and include tests with pull requests (`go test ./...`).

## License

[MIT](LICENSE)
