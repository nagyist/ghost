---
title: "ghost invite decline"
slug: "ghost_invite_decline"
description: "CLI reference for ghost invite decline"
---

## ghost invite decline

Decline an invitation

### Synopsis

Decline an invitation you've received.

The space is identified by its ID, as shown by 'ghost invite received'. By
default, you will be prompted to confirm, unless you use the --confirm flag.

```
ghost invite decline <space-id> [flags]
```

### Examples

```
  # Decline an invitation (with confirmation prompt)
  ghost invite decline x9y8z7w6v5

  # Decline without a confirmation prompt
  ghost invite decline x9y8z7w6v5 --confirm
```

### Options

```
      --confirm   Skip confirmation prompt
  -h, --help      help for decline
```

### Options inherited from parent commands

```
      --analytics           enable/disable usage analytics (default true)
      --color               enable colored output (default true)
      --config-dir string   config directory (default "~/.config/ghost")
      --version-check       check for updates (default true)
```

### SEE ALSO

* [ghost invite](ghost_invite.md)	 - Invite a user to the current space
