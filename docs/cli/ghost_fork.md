---
title: "ghost fork"
slug: "ghost_fork"
description: "CLI reference for ghost fork"
---

## ghost fork

Fork a database

### Synopsis

Fork an existing database to create a new independent copy.

To fork into an always-on dedicated database (not subject to space compute or
storage limits), use 'ghost fork-dedicated' instead.

```
ghost fork <name-or-id> [new-name] [flags]
```

### Examples

```
  # Fork a database with auto-generated name
  ghost fork my-database

  # Fork a database with a custom name
  ghost fork my-database myapp-experiment

  # Fork and output as JSON
  ghost fork my-database --json

  # Fork and output as YAML
  ghost fork my-database --yaml

  # Fork and wait for the database to be ready
  ghost fork my-database --wait
```

### Options

```
  -h, --help   help for fork
      --json   Output in JSON format
      --wait   Wait for the database to be ready before returning
      --yaml   Output in YAML format
```

### Options inherited from parent commands

```
      --analytics           enable/disable usage analytics (default true)
      --color               enable colored output (default true)
      --config-dir string   config directory (default "~/.config/ghost")
      --version-check       check for updates (default true)
```

### SEE ALSO

* [ghost](ghost.md)	 - CLI for managing Postgres databases
