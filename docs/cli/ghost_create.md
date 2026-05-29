---
title: "ghost create"
slug: "ghost_create"
description: "CLI reference for ghost create"
---

## ghost create

Create a new database

### Synopsis

Create a new Postgres database.

To create an always-on dedicated database (not subject to space compute or
storage limits), use 'ghost create-dedicated' instead.

```
ghost create [name] [flags]
```

### Examples

```
  # Create a database with auto-generated name
  ghost create

  # Create a database with a custom name
  ghost create myapp

  # Create a database from a share token
  ghost create myapp --from-share <token>

  # Create and output as JSON
  ghost create --json

  # Create and output as YAML
  ghost create --yaml

  # Create and wait for the database to be ready
  ghost create --wait
```

### Options

```
      --from-share string   Create the database from a share token
  -h, --help                help for create
      --json                Output in JSON format
      --wait                Wait for the database to be ready before returning
      --yaml                Output in YAML format
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
