---
title: "ghost invite"
slug: "ghost_invite"
description: "CLI reference for ghost invite"
---

## ghost invite

Invite a user to the current space

### Synopsis

Invite a user to the current space by email address.

The invitee accepts the invitation with their own Ghost account.

Roles:
  admin      Manage databases, members, and billing
  developer  Manage databases only (default)
  viewer     Read-only access

```
ghost invite <email> [flags]
```

### Examples

```
  # Invite a user as a developer (the default)
  ghost invite bob@example.com

  # Invite a user as an admin
  ghost invite bob@example.com --role admin

  # List invites you've sent
  ghost invite sent
```

### Options

```
  -h, --help          help for invite
      --json          Output in JSON format
      --role string   Role to grant the invitee (admin|developer|viewer) (default "developer")
      --yaml          Output in YAML format
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
* [ghost invite accept](ghost_invite_accept.md)	 - Accept an invitation
* [ghost invite cancel](ghost_invite_cancel.md)	 - Cancel an invite you've sent
* [ghost invite decline](ghost_invite_decline.md)	 - Decline an invitation
* [ghost invite list](ghost_invite_list.md)	 - List sent and received invites
* [ghost invite received](ghost_invite_received.md)	 - List invitations you've received
* [ghost invite sent](ghost_invite_sent.md)	 - List invites you've sent
