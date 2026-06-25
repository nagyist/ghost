---
title: "ghost space list"
slug: "ghost_space_list"
description: "CLI reference for ghost space list"
---

## ghost space list

List spaces

### Synopsis

List your Ghost spaces. The current space is marked with an asterisk.

```
ghost space list [flags]
```

### Examples

```
  # List your spaces
  ghost space list

  # Output as JSON
  ghost space list --json
```

### Options

```
  -h, --help   help for list
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

* [ghost space](ghost_space.md)	 - Show or manage spaces
