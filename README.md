# galaxy-update

> Update Ansible collection and role versions to latest

[![PyPI: Version](https://img.shields.io/pypi/v/galaxy-update?logo=pypi&logoColor=white)](https://pypi.org/project/galaxy-update)
[![GitHub: Release](https://img.shields.io/github/v/release/deadnews/galaxy-update?logo=github&logoColor=white)](https://github.com/deadnews/galaxy-update/releases/latest)
[![CI: Main](https://img.shields.io/github/actions/workflow/status/deadnews/galaxy-update/main.yml?branch=main&logo=github&logoColor=white&label=main)](https://github.com/deadnews/galaxy-update)
[![CI: Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/deadnews/galaxy-update/refs/heads/badges/coverage.json)](https://github.com/deadnews/galaxy-update)

## Installation

```sh
uv tool install galaxy-update
```

## Usage

```sh
Usage: galaxy-update [<files> ...] [flags]

Update Ansible collection and role versions to latest.

Flags:
  -v, --verbose    Show all entries, including current.
```

- When no files are given, `galaxy-update` auto-discovers
  `requirements.{yaml,yml}` under the current directory, recursively.
