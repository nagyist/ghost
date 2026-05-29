---
title: "ghost fork-dedicated"
slug: "ghost_fork-dedicated"
description: "CLI reference for ghost fork-dedicated"
---

## ghost fork-dedicated

Fork a database as dedicated

### Synopsis

Fork an existing database as a new dedicated instance. The fork inherits
the source database's data but runs as an always-on, billed instance.
A payment method must be on file.

Run 'ghost pricing' to see compute and storage pricing.

```
ghost fork-dedicated <name-or-id> [new-name] [flags]
```

### Examples

```
  # Fork as dedicated with default size (1x)
  ghost fork-dedicated my-database

  # Fork with a custom name
  ghost fork-dedicated my-database myapp-dedicated

  # Fork with a specific size
  ghost fork-dedicated my-database --size 4x

  # Fork and output as JSON
  ghost fork-dedicated my-database --json

  # Fork and wait for the database to be ready
  ghost fork-dedicated my-database --size 2x --wait
```

### Options

```
  -h, --help          help for fork-dedicated
      --json          Output in JSON format
      --size string   Database size (1x, 2x, 4x, 8x) (default "1x")
      --wait          Wait for the database to be ready before returning
      --yaml          Output in YAML format
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
