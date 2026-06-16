---
title: "ghost member list"
slug: "ghost_member_list"
description: "CLI reference for ghost member list"
---

## ghost member list

List space members

### Synopsis

List the members of the current space.

```
ghost member list [flags]
```

### Examples

```
  # List the members of the current space
  ghost member list

  # Output as JSON
  ghost member list --json
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

* [ghost member](ghost_member.md)	 - Manage space members
