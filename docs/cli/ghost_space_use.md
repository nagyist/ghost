---
title: "ghost space use"
slug: "ghost_space_use"
description: "CLI reference for ghost space use"
---

## ghost space use

Switch the current space

### Synopsis

Switch the current space.

Subsequent commands operate on the new current space. Run 'ghost space list'
to see your spaces and their IDs.

```
ghost space use <id> [flags]
```

### Examples

```
  # Switch the current space
  ghost space use x9y8z7w6v5
```

### Options

```
  -h, --help   help for use
```

### Options inherited from parent commands

```
      --analytics           enable/disable usage analytics (default true)
      --color               enable colored output (default true)
      --config-dir string   config directory (default "~/.config/ghost")
      --version-check       check for updates (default true)
```

### SEE ALSO

* [ghost space](ghost_space.md)	 - Show or manage spaces
