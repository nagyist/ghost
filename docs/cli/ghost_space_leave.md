---
title: "ghost space leave"
slug: "ghost_space_leave"
description: "CLI reference for ghost space leave"
---

## ghost space leave

Leave the current space

### Synopsis

Leave the current space, removing yourself from its members.

You cannot leave a space you own. By default, you will be prompted to
confirm, unless you use the --confirm flag.

```
ghost space leave [flags]
```

### Examples

```
  # Leave the current space (with confirmation prompt)
  ghost space leave

  # Leave without confirmation prompt
  ghost space leave --confirm
```

### Options

```
      --confirm   Skip confirmation prompt
  -h, --help      help for leave
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
