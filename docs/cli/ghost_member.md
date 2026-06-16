---
title: "ghost member"
slug: "ghost_member"
description: "CLI reference for ghost member"
---

## ghost member

Manage space members

### Synopsis

Manage the members of the current space.

Use 'ghost member list' to see the members of the current space,
'ghost member role' to change a member's role, and 'ghost member remove'
to remove a member from the space.

### Options

```
  -h, --help   help for member
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
* [ghost member list](ghost_member_list.md)	 - List space members
* [ghost member remove](ghost_member_remove.md)	 - Remove a member from the current space
* [ghost member role](ghost_member_role.md)	 - Change a member's role
