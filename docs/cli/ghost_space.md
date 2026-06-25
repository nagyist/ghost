---
title: "ghost space"
slug: "ghost_space"
description: "CLI reference for ghost space"
---

## ghost space

Show or manage spaces

### Synopsis

Show or manage Ghost spaces.

A space is a collection of databases with shared usage limits and billing.
The CLI operates on one space at a time — the current space. Running
'ghost space' with no subcommand shows details about the current space.

```
ghost space [flags]
```

### Examples

```
  # Show the current space
  ghost space

  # Output as JSON
  ghost space --json

  # Output as YAML
  ghost space --yaml
```

### Options

```
  -h, --help   help for space
      --json   Output in JSON format
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
* [ghost space list](ghost_space_list.md)	 - List spaces
* [ghost space rename](ghost_space_rename.md)	 - Rename the current space
* [ghost space use](ghost_space_use.md)	 - Switch the current space
