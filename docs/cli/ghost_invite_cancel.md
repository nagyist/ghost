---
title: "ghost invite cancel"
slug: "ghost_invite_cancel"
description: "CLI reference for ghost invite cancel"
---

## ghost invite cancel

Cancel an invite you've sent

### Synopsis

Cancel the pending invite sent to an email address for the current space.

By default, you will be prompted to confirm, unless you use the --confirm flag.

```
ghost invite cancel <email> [flags]
```

### Examples

```
  # Cancel an invite (with confirmation prompt)
  ghost invite cancel bob@example.com

  # Cancel an invite without a confirmation prompt
  ghost invite cancel bob@example.com --confirm
```

### Options

```
      --confirm   Skip confirmation prompt
  -h, --help      help for cancel
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
