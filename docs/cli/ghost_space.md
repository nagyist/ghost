---
title: "ghost space"
slug: "ghost_space"
description: "CLI reference for ghost space"
---

## ghost space

Manage spaces

### Synopsis

Manage Ghost spaces.

A space is a collection of databases with shared usage limits and billing.
The CLI operates on one space at a time — the current space. Use
'ghost space list' to see your spaces and 'ghost space use' to switch
between them.

### Options

```
  -h, --help   help for space
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
