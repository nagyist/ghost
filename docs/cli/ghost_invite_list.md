---
title: "ghost invite list"
slug: "ghost_invite_list"
description: "CLI reference for ghost invite list"
---

## ghost invite list

List sent and received invites

### Synopsis

List both the invites you've sent for the current space and the
invitations you've received across all spaces.

```
ghost invite list [flags]
```

### Examples

```
  # List sent and received invites
  ghost invite list

  # Output as JSON
  ghost invite list --json
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

* [ghost invite](ghost_invite.md)	 - Invite a user to the current space
