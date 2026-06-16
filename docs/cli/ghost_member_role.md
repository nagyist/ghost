---
title: "ghost member role"
slug: "ghost_member_role"
description: "CLI reference for ghost member role"
---

## ghost member role

Change a member's role

### Synopsis

Change a member's role in the current space.

Roles:
  admin      Manage databases, members, and billing
  developer  Manage databases only
  viewer     Read-only access

The owner role cannot be granted; every space has exactly one owner, and
the owner's role cannot be changed.

```
ghost member role <email> <admin|developer|viewer> [flags]
```

### Examples

```
  # Make a member an admin
  ghost member role bob@example.com admin

  # Restrict a member to read-only access
  ghost member role bob@example.com viewer
```

### Options

```
  -h, --help   help for role
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
