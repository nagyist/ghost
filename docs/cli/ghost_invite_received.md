---
title: "ghost invite received"
slug: "ghost_invite_received"
description: "CLI reference for ghost invite received"
---

## ghost invite received

List invitations you've received

### Synopsis

List the pending invitations you've received across all spaces.

Accept one with 'ghost invite accept <space-id>' or decline it with
'ghost invite decline <space-id>'.

```
ghost invite received [flags]
```

### Examples

```
  # List invitations you've received
  ghost invite received

  # Output as JSON
  ghost invite received --json
```

### Options

```
  -h, --help   help for received
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
