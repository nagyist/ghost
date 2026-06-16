---
title: "ghost invite sent"
slug: "ghost_invite_sent"
description: "CLI reference for ghost invite sent"
---

## ghost invite sent

List invites you've sent

### Synopsis

List the pending invites you've sent for the current space.

```
ghost invite sent [flags]
```

### Examples

```
  # List invites you've sent
  ghost invite sent

  # Output as JSON
  ghost invite sent --json
```

### Options

```
  -h, --help   help for sent
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

* [ghost invite](ghost_invite.md)	 - Invite a user to the current space
