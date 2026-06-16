---
title: "ghost member remove"
slug: "ghost_member_remove"
description: "CLI reference for ghost member remove"
---

## ghost member remove

Remove a member from the current space

### Synopsis

Remove a member from the current space.

The space owner cannot be removed. By default, you will be prompted to
confirm the removal, unless you use the --confirm flag.

```
ghost member remove <email> [flags]
```

### Examples

```
  # Remove a member (with confirmation prompt)
  ghost member remove bob@example.com

  # Remove a member without confirmation prompt
  ghost member remove bob@example.com --confirm
```

### Options

```
      --confirm   Skip confirmation prompt
  -h, --help      help for remove
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
